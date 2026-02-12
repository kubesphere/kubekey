package connector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	_const "github.com/kubesphere/kubekey/v4/pkg/const"
)

const (
	// gatherFactsCacheJSON indicates that facts should be cached in JSON format
	gatherFactsCacheJSON = "jsonfile"
	// gatherFactsCacheYAML indicates that facts should be cached in YAML format
	gatherFactsCacheYAML = "yamlfile"
	// gatherFactsCacheMemory indicates that facts should be cached in memory
	gatherFactsCacheMemory = "memory"
)

var cache = &memoryCache{
	cache: make(map[string]map[string]any),
}

type memoryCache struct {
	cache      map[string]map[string]any
	cacheMutex sync.RWMutex
}

// Get retrieves cached data for a host (thread-safe).
func (m *memoryCache) Get(hostname string) (map[string]any, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()
	data, exists := m.cache[hostname]
	return data, exists
}

// Set stores data for a host (thread-safe).
func (m *memoryCache) Set(hostname string, data map[string]any) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.cache[hostname] = data
}

// Delete removes the cached data for a specific hostname (thread-safe).
func (m *memoryCache) Delete(hostname string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	delete(m.cache, hostname)
}

// GatherFacts defines an interface for retrieving host information
type GatherFacts interface {
	// HostInfo returns a map of host facts gathered from the system
	HostInfo(ctx context.Context) (map[string]any, error)
}

// cacheGatherFact implements GatherFacts with caching capabilities
type cacheGatherFact struct {
	// inventoryName is the name of the host in the inventory
	inventoryName string
	// cacheType specifies the format to cache facts (json, yaml, or memory)
	cacheType string
	// cacheDir is the cache dir in local
	cacheDir string
	// getHostInfoFn is the function that actually gathers host information
	getHostInfoFn func(context.Context) (map[string]any, error)
}

// newCacheGatherFact creates a new cacheGatherFact instance
func newCacheGatherFact(inventoryName, cacheType, workdir string, getHostInfoFn func(context.Context) (map[string]any, error)) *cacheGatherFact {
	return &cacheGatherFact{
		inventoryName: inventoryName,
		cacheType:     cacheType,
		cacheDir:      filepath.Join(workdir, _const.RuntimeDir, _const.RuntimeGatherFactsCacheDir),
		getHostInfoFn: getHostInfoFn,
	}
}

// HostInfo returns host information from cache or fetches it remotely if not cached.
// The caching behavior depends on the configured cache type (JSON, YAML, or memory).
func (c *cacheGatherFact) HostInfo(ctx context.Context) (map[string]any, error) {
	switch c.cacheType {
	case gatherFactsCacheJSON:
		return c.handleJSONCache(ctx)
	case gatherFactsCacheYAML:
		return c.handleYAMLCache(ctx)
	case gatherFactsCacheMemory:
		return c.handleMemoryCache(ctx)
	default:
		// fallback: delete possible cache and fetch directly
		_ = os.Remove(filepath.Join(c.cacheDir, c.inventoryName+".json"))
		_ = os.Remove(filepath.Join(c.cacheDir, c.inventoryName+".yaml"))
		cache.Delete(c.inventoryName)
		return c.getHostInfoFn(ctx)
	}
}

// ensureCacheDir ensures the cache directory exists, creating it if necessary
func (c *cacheGatherFact) ensureCacheDir() error {
	if _, err := os.Stat(c.cacheDir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(c.cacheDir, os.ModePerm)
		}
		return err
	}
	return nil
}

// handleJSONCache handles caching host information in JSON format.
// It attempts to read from the cache file first, falling back to remote fetch if needed.
func (c *cacheGatherFact) handleJSONCache(ctx context.Context) (map[string]any, error) {
	if err := c.ensureCacheDir(); err != nil {
		return nil, errors.Wrapf(err, "json cache dir error for host %q", c.inventoryName)
	}
	filename := filepath.Join(c.cacheDir, c.inventoryName+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.V(4).InfoS("json cache miss. fetching remotely.", "filename", filename)
		return c.fetchAndCache(ctx, filename, json.Marshal)
	}
	var result map[string]any
	return result, json.Unmarshal(data, &result)
}

// handleYAMLCache handles caching host information in YAML format.
// It attempts to read from the cache file first, falling back to remote fetch if needed.
func (c *cacheGatherFact) handleYAMLCache(ctx context.Context) (map[string]any, error) {
	if err := c.ensureCacheDir(); err != nil {
		return nil, errors.Wrapf(err, "yaml cache dir error for host %q", c.inventoryName)
	}
	filename := filepath.Join(c.cacheDir, c.inventoryName+".yaml")
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.V(4).InfoS("yaml cache miss. fetching remotely.", "filename", filename)
		return c.fetchAndCache(ctx, filename, yaml.Marshal)
	}
	var result map[string]any
	return result, yaml.Unmarshal(data, &result)
}

// fetchAndCache fetches host information remotely and caches it to a file.
// marshalFn specifies how to marshal the data (JSON or YAML).
func (c *cacheGatherFact) fetchAndCache(
	ctx context.Context,
	filename string,
	marshalFn func(any) ([]byte, error),
) (map[string]any, error) {
	hostInfo, err := c.getHostInfoFn(ctx)
	if err != nil {
		return nil, err
	}
	data, err := marshalFn(hostInfo)
	if err != nil {
		return nil, err
	}
	return hostInfo, os.WriteFile(filename, data, os.ModePerm)
}

// handleMemoryCache handles caching host information in memory.
// It checks the in-memory cache first, falling back to remote fetch if needed.
func (c *cacheGatherFact) handleMemoryCache(ctx context.Context) (map[string]any, error) {
	if cached, exists := cache.Get(c.inventoryName); exists {
		return cached, nil
	}
	hostInfo, err := c.getHostInfoFn(ctx)
	if err != nil {
		return nil, err
	}
	cache.Set(c.inventoryName, hostInfo)
	return hostInfo, nil
}
