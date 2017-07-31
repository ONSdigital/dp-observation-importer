package dimension

// IDCache is an in memory cache of dimensions with database id's.
type DimensionMemoryCache struct {
	idStore IDStore
	memoryCache map[string]map[string]string // instanceID > dimensionName > nodeId
}

// IDStore represents the data store for dimension id's
type IDStore interface {
	GetIDs(instanceID string) (map[string]string, error)
}

// NewIDCache returns a new cache instance that uses the given data store.
func NewIDCache(idStore IDStore) *DimensionMemoryCache {
	return &DimensionMemoryCache{
		idStore:idStore,
		memoryCache: make(map[string]map[string]string),
	}
}

// GetIDs returns all dimensions for a given instanceID
func (cache *DimensionMemoryCache) GetNodeIDs(instanceID string) (map[string]string, error) {

	dimensions, ok := cache.memoryCache[instanceID]

	if ok {

		return dimensions, nil
	}

	newDimensions, storeError := cache.idStore.GetIDs(instanceID)
	if storeError != nil {
		return nil, storeError
	}

	cache.memoryCache[instanceID] = newDimensions

	return newDimensions, nil
}
