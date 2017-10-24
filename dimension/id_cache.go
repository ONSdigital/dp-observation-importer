package dimension

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

// MemoryCache is an in memory cache of dimensions with database id's.
type MemoryCache struct {
	store       IDStore
	memoryCache *cache.Cache // instanceID > dimensionName > nodeId
}

// IDStore represents the data store for dimension id's
type IDStore interface {
	GetIDs(instanceID string) (map[string]string, error)
}

// NewIDCache returns a new cache instance that uses the given data store.
func NewIDCache(idStore IDStore, cacheTTL time.Duration) *MemoryCache {
	return &MemoryCache{
		store:       idStore,
		memoryCache: cache.New(cacheTTL, 15*time.Minute),
	}
}

// GetNodeIDs returns all dimensions for a given instanceID
func (mc *MemoryCache) GetNodeIDs(instanceID string) (map[string]string, error) {

	if dimensions, ok := mc.memoryCache.Get(instanceID); ok {
		return dimensions.(map[string]string), nil
	}

	newDimensions, err := mc.store.GetIDs(instanceID)
	if err != nil {
		return nil, err
	}

	err = mc.memoryCache.Add(instanceID, newDimensions, cache.DefaultExpiration)
	if err != nil {
		return nil, err
	}

	return newDimensions, nil
}
