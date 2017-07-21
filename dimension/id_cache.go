package dimension

// IDCache is an in memory cache of dimensions with database id's.
type IDCache struct {
	idStore IDStore
}

// IDStore represents the data store for dimension id's
type IDStore interface {
	GetIDs(instanceID string) ([]*Dimension, error)
}

// NewIDCache returns a new cache instance that uses the given data store.
func NewIDCache(idStore IDStore) *IDCache {
	return &IDCache{
		idStore:idStore,
	}
}

// GetIDs returns all dimensions for a given instanceID
func (cache *IDCache) Get(instanceID string) ([]*Dimension, error) {

	// todo: implement in memory cache

	return cache.idStore.GetIDs(instanceID)
}
