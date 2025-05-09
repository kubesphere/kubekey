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

// HostInfo retrieves host facts, either from cache or by gathering them directly
func (c *cacheGatherFact) HostInfo(ctx context.Context) (map[string]any, error) {
	switch c.cacheType {
	case gatherFactsCacheJSON:
		if _, err := os.Stat(c.cacheDir); err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "failed to stat local dir %q for host %q", c.cacheDir, c.inventoryName)
			}
			// if dir is not exist, create it.
			if err := os.MkdirAll(c.cacheDir, os.ModePerm); err != nil {
				return nil, errors.Wrapf(err, "failed to create local dir %q for host %q", c.cacheDir, c.inventoryName)
			}
		}
		// os.MkdirAll(, fs.ModePerm)
		// Build path for JSON cache file
		filename := filepath.Join(c.cacheDir, c.inventoryName+".json")
		data, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).Infof("cannot get cache file from %q. get from remote", filename)
			// Cache miss - gather facts directly
			hostInfo, err := c.getHostInfoFn(ctx)
			if err != nil {
				return nil, err
			}
			// Store gathered facts in JSON cache
			cacheData, err := json.Marshal(hostInfo)
			if err != nil {
				return nil, err
			}
			return hostInfo, os.WriteFile(filename, cacheData, os.ModePerm)
		}
		// Cache hit - unmarshal and return cached data
		var result map[string]any
		return result, json.Unmarshal(data, &result)
	case gatherFactsCacheYAML:
		if _, err := os.Stat(c.cacheDir); err != nil {
			if !os.IsNotExist(err) {
				return nil, errors.Wrapf(err, "failed to stat local dir %q for host %q", c.cacheDir, c.inventoryName)
			}
			// if dir is not exist, create it.
			if err := os.MkdirAll(c.cacheDir, os.ModePerm); err != nil {
				return nil, errors.Wrapf(err, "failed to create local dir %q for host %q", c.cacheDir, c.inventoryName)
			}
		}
		// Build path for YAML cache file
		filename := filepath.Join(c.cacheDir, c.inventoryName+".yaml")
		data, err := os.ReadFile(filename)
		if err != nil {
			klog.V(4).Infof("cannot get cache file from %q. get from remote", filename)
			// Cache miss - gather facts directly
			hostInfo, err := c.getHostInfoFn(ctx)
			if err != nil {
				return nil, err
			}
			// Store gathered facts in YAML cache
			cacheData, err := yaml.Marshal(hostInfo)
			if err != nil {
				return nil, err
			}
			return hostInfo, os.WriteFile(filename, cacheData, os.ModePerm)
		}
		// Cache hit - unmarshal and return cached data
		var result map[string]any
		return result, yaml.Unmarshal(data, &result)
	case gatherFactsCacheMemory:
		if cached, exists := cache.Get(c.inventoryName); exists {
			return cached, nil
		}
		hostInfo, err := c.getHostInfoFn(ctx)
		if err != nil {
			return nil, err
		}
		cache.Set(c.inventoryName, hostInfo)
		return hostInfo, nil
	default: // don't get from cache
		// Clear cache before re-fetching
		switch c.cacheType {
		case gatherFactsCacheJSON:
			_ = os.Remove(filepath.Join(c.cacheDir, c.inventoryName+".json"))
		case gatherFactsCacheYAML:
			_ = os.Remove(filepath.Join(c.cacheDir, c.inventoryName+".yaml"))
		case gatherFactsCacheMemory:
			cache.Delete(c.inventoryName)
		}
		// todo clear cache
		return c.getHostInfoFn(ctx)
	}
}
