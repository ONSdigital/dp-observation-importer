package dimension

import (
	"github.com/patrickmn/go-cache"
	"time"
)

// IDCache is an in memory cache of dimensions with database id's.
type DimensionMemoryCache struct {
	idStore IDStore
	memoryCache *cache.Cache // instanceID > dimensionName > nodeId
}

// IDStore represents the data store for dimension id's
type IDStore interface {
	GetIDs(instanceID string) (map[string]string, error)
}

// NewIDCache returns a new cache instance that uses the given data store.
func NewIDCache(idStore IDStore, cacheTTL time.Duration) *DimensionMemoryCache {
	return &DimensionMemoryCache{
		idStore:idStore,
		memoryCache: cache.New(cacheTTL, 15 * time.Minute),
	}
}

// GetIDs returns all dimensions for a given instanceID
func (dmc *DimensionMemoryCache) GetNodeIDs(instanceID string) (map[string]string, error) {

	dimensions, ok := dmc.memoryCache.Get(instanceID)

	if ok {

		return dimensions.(map[string]string), nil
	}

	newDimensions, storeError := dmc.idStore.GetIDs(instanceID)
	if storeError != nil {
		return nil, storeError
	}

	dmc.memoryCache.Add(instanceID, newDimensions, cache.DefaultExpiration)

	return newDimensions, nil
}
