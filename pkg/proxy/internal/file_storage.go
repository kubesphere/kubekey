package internal

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	apistorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/storagebackend/factory"
)

const (
	// dataFileSuffix is the suffix for data files
	dataFileSuffix = ".yaml"
	// revisionFile is the file name for global revision number
	revisionFile = "_revision"
	// deleteTagSuffix is the suffix for deletion marker files
	deleteTagSuffix = "-deleted"
)

// fileStore implements the apistorage.Interface interface for local file storage
type fileStore struct {
	// rootDir is the root directory for storage
	rootDir string
	// resourcePrefix is the resource prefix (group/version/resource)
	resourcePrefix string
	// clusterScoped indicates if the resource is cluster-scoped
	clusterScoped bool
	// codec is the encoder/decoder
	codec runtime.Codec
	// versioner manages resource versions
	versioner apistorage.Versioner
	// currentRev is the current revision number (globally incrementing)
	revisionMux sync.RWMutex
	currentRev  uint64
	// newFunc is the function to create new objects
	newFunc func() runtime.Object
}

// Ensure fileStore implements apistorage.Interface
var _ apistorage.Interface = &fileStore{}

// newFileStorage creates a new local file storage
func newFileStorage(rootDir string, resourcePrefix string, groupResource schema.GroupResource, codec runtime.Codec, newFunc func() runtime.Object, isClusterScoped bool) (*fileStore, factory.DestroyFunc, error) {
	s := &fileStore{
		rootDir:        rootDir,
		resourcePrefix: resourcePrefix,
		clusterScoped:  isClusterScoped,
		codec:          codec,
		versioner:      apistorage.APIObjectVersioner{},
		currentRev:     1, // Start from 1 to avoid "illegal resource version from storage: 0"
		newFunc:        newFunc,
	}

	if err := s.loadRevision(); err != nil {
		return nil, nil, err
	}

	return s, func() {}, nil
}

// ================================================
// Version Management
// ================================================

// loadRevision loads the revision number from file
func (s *fileStore) loadRevision() error {
	data, err := os.ReadFile(filepath.Join(s.rootDir, s.resourcePrefix, revisionFile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.Wrap(err, "failed to read revision file")
	}
	s.currentRev, err = strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to parse revision")
	}
	return nil
}

// updateRevision writes the global revision number to file, incrementing it atomically
func (s *fileStore) updateRevision() error {
	s.revisionMux.Lock()
	s.currentRev++
	rev := s.currentRev
	s.revisionMux.Unlock()

	revisionPath := filepath.Join(s.rootDir, s.resourcePrefix, revisionFile)
	return os.WriteFile(revisionPath, []byte(strconv.FormatUint(rev, 10)), 0644)
}

// ================================================
// Key Preparation
// ================================================

// prepareKey returns the storage key as a full path based on rootDir
func (s *fileStore) prepareKey(key string) string {
	// Remove leading separator from path
	return filepath.Join(s.rootDir, key)
}

// getKeyFromPath extracts the storage key from an absolute path (used for Watch)
func (s *fileStore) getKeyFromPath(absPath string) string {
	relPath := strings.TrimPrefix(absPath, s.rootDir)
	// Ensure path starts with separator
	if !strings.HasPrefix(relPath, string(filepath.Separator)) {
		relPath = string(filepath.Separator) + relPath
	}
	return s.resourcePrefix + relPath
}

// ================================================
// File Operations
// ================================================

// writeObject writes an object to file
func (s *fileStore) writeObject(key string, obj runtime.Object) error {
	filePath := key + dataFileSuffix
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return errors.Wrap(err, "failed to create directory")
	}
	data, err := runtime.Encode(s.codec, obj)
	if err != nil {
		return errors.Wrap(err, "failed to encode object")
	}
	return os.WriteFile(filePath, data, 0644)
}

// readObject reads an object from a file
func (s *fileStore) readObject(key string, obj runtime.Object) error {
	data, err := os.ReadFile(key + dataFileSuffix)
	if err != nil {
		return err
	}
	return runtime.DecodeInto(s.codec, data, obj)
}

// walkDirectory recursively traverses a directory, invoking fn for each valid object
func (s *fileStore) walkDirectory(prefix string, fn func(path string, obj runtime.Object) error) error {
	return filepath.WalkDir(prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Only process .yaml files
		if !strings.HasSuffix(path, dataFileSuffix) {
			return nil
		}
		// Skip deleted marker files
		if strings.HasSuffix(path, deleteTagSuffix) {
			return nil
		}
		// Skip revision files
		if strings.HasSuffix(path, revisionFile) {
			return nil
		}

		// Infer object type from path and read
		key := strings.TrimSuffix(path, dataFileSuffix)
		obj := s.newFunc()
		if err := s.readObject(key, obj); err != nil {
			return fn(path, nil)
		}
		return fn(path, obj)
	})
}

// ================================================
// CRUD Operations
// ================================================

// Create stores a new object at the given key. Fails if key already exists.
func (s *fileStore) Create(ctx context.Context, key string, obj, out runtime.Object, ttl uint64) error {
	preparedKey := s.prepareKey(key)

	// Check for resource version; must not be set on create
	if version, err := s.versioner.ObjectResourceVersion(obj); err == nil && version != 0 {
		return apistorage.ErrResourceVersionSetOnCreate
	}

	// Set initial resource version
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return errors.Wrap(err, "failed to get accessor")
	}
	metaObj.SetResourceVersion("1")

	// Check if file already exists
	filePath := preparedKey + dataFileSuffix
	if _, err := os.Stat(filePath); err == nil {
		return apistorage.NewKeyExistsError(preparedKey, 0)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// Write object to file
	if err := s.writeObject(preparedKey, obj); err != nil {
		return err
	}

	// Update global revision number
	if err := s.updateRevision(); err != nil {
		return err
	}

	// Optionally, read out the created resource into 'out'
	if out != nil {
		if err := s.readObject(preparedKey, out); err != nil {
			return err
		}
	}

	return nil
}

// Get fetches an object from storage by key
func (s *fileStore) Get(ctx context.Context, key string, opts apistorage.GetOptions, out runtime.Object) error {
	preparedKey := s.prepareKey(key)

	// Read object data
	if err := s.readObject(preparedKey, out); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if opts.IgnoreNotFound {
				return runtime.SetZeroValue(out)
			}
			return apistorage.NewKeyNotFoundError(preparedKey, 0)
		}
		return err
	}

	// Validate resource version if provided
	if opts.ResourceVersion != "" {
		rv, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
		if err != nil {
			return apistorage.NewInvalidObjError(preparedKey, err.Error())
		}
		currentRV, _ := s.versioner.ObjectResourceVersion(out)
		if uint64(currentRV) < rv {
			return apistorage.NewInvalidObjError(preparedKey, "resource version too old")
		}
	}

	return nil
}

// GetList fetches all objects under a storage key, supporting pagination and filtering
func (s *fileStore) GetList(ctx context.Context, key string, opts apistorage.ListOptions, listObj runtime.Object) error {
	listPtr, err := meta.GetItemsPtr(listObj)
	if err != nil {
		return errors.Wrap(err, "failed to get items ptr")
	}

	v, err := conversion.EnforcePtr(listPtr)
	if err != nil || v.Kind() != reflect.Slice {
		return errors.New("need ptr to slice")
	}

	// Build matching rules for resource version and continue key
	resourceVersionMatchRule, continueKeyMatchRule, err := s.buildMatchRules(key, opts)
	if err != nil {
		return err
	}

	// Get root directory entries (note: key is relative and needs to be translated)
	rootEntries, isAllNamespace, err := s.getRootEntries(key)
	if err != nil {
		return errors.Wrap(err, "failed to get root entries")
	}

	// Convert key to absolute path for file operations
	preparedKey := s.prepareKey(key)

	var lastKey string
	var hasMore bool

	// Iterate entries
	for i, entry := range rootEntries {
		if isAllNamespace {
			err = s.processNamespaceDirectory(preparedKey, entry, v, continueKeyMatchRule, resourceVersionMatchRule, &lastKey, &hasMore, opts, listObj)
		} else {
			err = s.processResourceFile(preparedKey, entry, v, continueKeyMatchRule, resourceVersionMatchRule, &lastKey, opts, listObj)
		}
		if err != nil {
			return err
		}

		// Check if limit is reached
		if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
			hasMore = i != len(rootEntries)-1
			break
		}
	}

	return s.handleResult(listObj, v, lastKey, hasMore)
}

// buildMatchRules returns filter functions for resource version and for continue key
func (s *fileStore) buildMatchRules(key string, opts apistorage.ListOptions) (func(uint64) bool, func(string) bool, error) {
	resourceVersionMatchRule := func(uint64) bool { return true }
	continueKeyMatchRule := func(path string) bool {
		// Accept only YAML data files, not delete or revision files
		return strings.HasSuffix(path, dataFileSuffix) &&
			!strings.HasSuffix(path, deleteTagSuffix) &&
			!strings.HasSuffix(path, revisionFile)
	}

	switch {
	case opts.Recursive && opts.Predicate.Continue != "":
		// Handle continue token
		continueKey, _, err := apistorage.DecodeContinue(opts.Predicate.Continue, key)
		if err != nil {
			return nil, nil, errors.Wrap(err, "invalid continue token")
		}

		var startRead bool
		continueKeyMatchRule = func(path string) bool {
			if path == continueKey {
				startRead = true
			}
			return startRead && path != continueKey
		}

	case opts.ResourceVersion != "":
		// Handle resource version matching
		parsedRV, err := s.versioner.ParseResourceVersion(opts.ResourceVersion)
		if err != nil {
			return nil, nil, errors.Wrap(err, "invalid resource version")
		}

		switch opts.ResourceVersionMatch {
		case metav1.ResourceVersionMatchNotOlderThan:
			resourceVersionMatchRule = func(u uint64) bool { return u >= parsedRV }
		case metav1.ResourceVersionMatchExact:
			resourceVersionMatchRule = func(u uint64) bool { return u == parsedRV }
		case "":
			// Compatible with older versions, match all
		default:
			return nil, nil, errors.New("unknown ResourceVersionMatch")
		}
	}

	return resourceVersionMatchRule, continueKeyMatchRule, nil
}

// getRootEntries gets root directory entries
// key can be in one of two forms:
//   - "apis/apps/deployments" (with resourcePrefix prefix)
//   - "" or "/deployments" (when listing all resources in GetList)
func (s *fileStore) getRootEntries(key string) ([]os.DirEntry, bool, error) {
	var allNamespace bool

	// Extract relative path (remove resourcePrefix prefix)
	relKey := strings.TrimPrefix(key, s.resourcePrefix)
	relKey = strings.TrimPrefix(relKey, string(filepath.Separator))

	// Determine if listing all namespaces based on relative path and resource scope
	// For cluster-scoped resources, allNamespace is always false (no namespace directories)
	// For namespace-scoped resources, if relative path is empty or only separator, list all namespaces
	if !s.clusterScoped {
		switch relKey {
		case "", string(filepath.Separator):
			allNamespace = true
		default:
			allNamespace = false
		}
	}
	// For cluster-scoped resources, allNamespace remains false

	// Actual path
	preparedKey := s.prepareKey(key)
	if _, err := os.Stat(preparedKey); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(preparedKey, 0755); err != nil {
				return nil, allNamespace, errors.Wrap(err, "failed to create directory")
			}
		} else {
			return nil, allNamespace, errors.Wrap(err, "failed to stat directory")
		}
	}

	entries, err := os.ReadDir(preparedKey)
	return entries, allNamespace, err
}

// processNamespaceDirectory processes namespace directories or cluster-scoped resources
func (s *fileStore) processNamespaceDirectory(key string, entry os.DirEntry, v reflect.Value, continueKeyMatchRule func(string) bool, resourceVersionMatchRule func(uint64) bool, lastKey *string, hasMore *bool, opts apistorage.ListOptions, listObj runtime.Object) error {
	entryPath := filepath.Join(key, entry.Name())

	if entry.IsDir() {
		// Namespace-scoped resources: key/<namespace>/
		nsDir := entryPath
		entries, err := os.ReadDir(nsDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return errors.Wrap(err, "failed to read namespace directory")
		}

		for _, e := range entries {
			err := s.processResourceFile(nsDir, e, v, continueKeyMatchRule, resourceVersionMatchRule, lastKey, opts, listObj)
			if err != nil {
				return err
			}

			if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
				*hasMore = true
				return nil
			}
		}
	} else {
		// Cluster-scoped resources: key/<resource>.yaml (directly under key)
		err := s.processResourceFile(key, entry, v, continueKeyMatchRule, resourceVersionMatchRule, lastKey, opts, listObj)
		if err != nil {
			return err
		}

		if opts.Predicate.Limit != 0 && int64(v.Len()) >= opts.Predicate.Limit {
			*hasMore = true
		}
	}

	return nil
}

// processResourceFile processes a single resource file
func (s *fileStore) processResourceFile(parentDir string, entry os.DirEntry, v reflect.Value, continueKeyMatchRule func(string) bool, resourceVersionMatchRule func(uint64) bool, lastKey *string, opts apistorage.ListOptions, listObj runtime.Object) error {
	if entry.IsDir() {
		return nil
	}

	currentKey := filepath.Join(parentDir, entry.Name())
	if !continueKeyMatchRule(currentKey) {
		return nil
	}

	data, err := os.ReadFile(currentKey)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	obj := s.newFunc()
	if err := runtime.DecodeInto(s.codec, data, obj); err != nil {
		return errors.Wrap(err, "failed to decode object")
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return errors.Wrap(err, "failed to get accessor")
	}

	rv, err := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())
	if err != nil {
		return errors.Wrap(err, "failed to parse resource version")
	}

	// Resource version matching
	if !resourceVersionMatchRule(rv) {
		return nil
	}

	// Filter condition matching
	if matched, err := opts.Predicate.Matches(obj); err != nil {
		return err
	} else if matched {
		v.Set(reflect.Append(v, reflect.ValueOf(obj).Elem()))
		// lastKey needs to be converted to format with resourcePrefix for continue token
		*lastKey = s.getKeyFromPath(currentKey)
	}

	return nil
}

// handleResult handles the final result
func (s *fileStore) handleResult(listObj runtime.Object, v reflect.Value, lastKey string, hasMore bool) error {
	if v.IsNil() {
		v.Set(reflect.MakeSlice(v.Type(), 0, 0))
	}

	if hasMore {
		// Generate continue token
		next, err := apistorage.EncodeContinue(lastKey+"\x00", "", 0)
		if err != nil {
			return errors.Wrap(err, "failed to encode continue")
		}
		return s.versioner.UpdateList(listObj, s.currentRev, next, nil)
	}

	return s.versioner.UpdateList(listObj, s.currentRev, "", nil)
}

// GuaranteedUpdate implements guaranteed update
func (s *fileStore) GuaranteedUpdate(
	ctx context.Context, key string, destination runtime.Object, ignoreNotFound bool,
	preconditions *apistorage.Preconditions, tryUpdate apistorage.UpdateFunc, cachedExistingObject runtime.Object) error {
	preparedKey := s.prepareKey(key)

	_, err := conversion.EnforcePtr(destination)
	if err != nil {
		return errors.Wrap(err, "unable to convert output object to pointer")
	}

	// Get current object
	var currentObj runtime.Object
	if cachedExistingObject != nil {
		currentObj = cachedExistingObject
	} else {
		currentObj = s.newFunc()
		if err := s.readObject(preparedKey, currentObj); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if ignoreNotFound {
					return runtime.SetZeroValue(destination)
				}
				return apistorage.NewKeyNotFoundError(preparedKey, 0)
			}
			return err
		}
	}

	// Check preconditions
	if preconditions != nil {
		if err := preconditions.Check(preparedKey, currentObj); err != nil {
			return err
		}
	}

	// Get current version
	metaObj, err := meta.Accessor(currentObj)
	if err != nil {
		return errors.Wrap(err, "failed to get accessor")
	}
	oldVersion, _ := s.versioner.ParseResourceVersion(metaObj.GetResourceVersion())

	// Perform update
	ret, _, err := tryUpdate(currentObj, apistorage.ResponseMeta{ResourceVersion: oldVersion + 1})
	if err != nil {
		return err
	}

	// Update resource version
	metaObj, _ = meta.Accessor(ret)
	metaObj.SetResourceVersion(strconv.FormatUint(oldVersion+1, 10))

	// Check if there are changes
	currentData, _ := runtime.Encode(s.codec, currentObj)
	newData, _ := runtime.Encode(s.codec, ret)
	if string(currentData) != string(newData) {
		if err := s.writeObject(preparedKey, ret); err != nil {
			return err
		}
		// Update global revision number
		if err := s.updateRevision(); err != nil {
			return err
		}
	}

	return s.readObject(preparedKey, destination)
}

// Delete implements the delete operation
func (s *fileStore) Delete(
	ctx context.Context, key string, out runtime.Object, preconditions *apistorage.Preconditions,
	validateDeletion apistorage.ValidateObjectFunc, cachedExistingObject runtime.Object, opts apistorage.DeleteOptions) error {
	preparedKey := s.prepareKey(key)

	// Get current object
	var currentObj runtime.Object
	if cachedExistingObject != nil {
		currentObj = cachedExistingObject
	} else {
		currentObj = s.newFunc()
		if err := s.readObject(preparedKey, currentObj); err != nil {
			return err
		}
	}

	// Check preconditions
	if preconditions != nil {
		if err := preconditions.Check(preparedKey, currentObj); err != nil {
			return err
		}
	}

	// Validate deletion
	if validateDeletion != nil {
		if err := validateDeletion(ctx, currentObj); err != nil {
			return err
		}
	}

	// Rename file to trigger watcher deletion mechanism
	deletedFilePath := preparedKey + dataFileSuffix + deleteTagSuffix
	if err := os.Rename(preparedKey+dataFileSuffix, deletedFilePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return apistorage.NewKeyNotFoundError(preparedKey, 0)
		}
		return errors.Wrap(err, "failed to rename to deleted file")
	}

	// Update global revision number
	if err := s.updateRevision(); err != nil {
		return err
	}

	return nil
}

// ================================================
// Watch Implementation
// ================================================

// Watch implements the watch operation
func (s *fileStore) Watch(ctx context.Context, key string, opts apistorage.ListOptions) (watch.Interface, error) {
	preparedKey := s.prepareKey(key)
	return newFileWatcher(s.rootDir, preparedKey, s.codec, s.newFunc, opts, s.clusterScoped)
}

// ================================================
// Other Interface Methods
// ================================================

// Versioner returns the versioner
func (s *fileStore) Versioner() apistorage.Versioner {
	return s.versioner
}

// CompactRevision returns the compact revision
func (s *fileStore) CompactRevision() int64 {
	return 0
}

// Stats returns storage statistics
func (s *fileStore) Stats(ctx context.Context) (apistorage.Stats, error) {
	count := int64(0)

	if err := s.walkDirectory(s.rootDir, func(path string, obj runtime.Object) error {
		if obj == nil {
			return nil
		}
		count++
		return nil
	}); err != nil {
		return apistorage.Stats{}, errors.Wrap(err, "failed to walk directory")
	}

	return apistorage.Stats{
		ObjectCount: count,
	}, nil
}

// ReadinessCheck checks if storage is ready
func (s *fileStore) ReadinessCheck() error {
	_, err := os.Stat(s.rootDir)
	return err
}

// RequestWatchProgress requests watch progress
func (s *fileStore) RequestWatchProgress(ctx context.Context) error {
	return nil
}

// EnableResourceSizeEstimation enables resource size estimation
func (s *fileStore) EnableResourceSizeEstimation(keysFunc apistorage.KeysFunc) error {
	return nil
}

// GetCurrentResourceVersion returns the current resource version
func (s *fileStore) GetCurrentResourceVersion(ctx context.Context) (uint64, error) {
	s.revisionMux.RLock()
	defer s.revisionMux.RUnlock()
	// currentRev is already uint64 and always starts from 1
	return s.currentRev, nil
}

// Close closes the storage
func (s *fileStore) Close() {
	// fileStore does not need close operation
}
