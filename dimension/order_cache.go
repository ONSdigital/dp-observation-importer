package dimension

// OrderCache is cache the for order of dimensions in the input file
type HeaderCache struct {
	orderStore OrderStore
	memoryCache map[string][]string
}

// OrderStore represents the data store for dimension order
type OrderStore interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewOrderCache returns a new instance of the order cache that uses the given OrderStore.
func NewOrderCache(orderStore OrderStore) *HeaderCache {
	return &HeaderCache{
		orderStore:orderStore,
		memoryCache: make(map[string] []string),
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (cache *HeaderCache) GetOrder(instanceID string) ([]string, error) {
	headers, ok := cache.memoryCache[instanceID]

	if ok {
		return headers, nil
	}

	newHeaders, storeError := cache.orderStore.GetOrder(instanceID)
	if storeError != nil {
		return nil, storeError
	}
	// TODO Time to live on cached items
	cache.memoryCache[instanceID] = newHeaders

	return newHeaders, nil
}
