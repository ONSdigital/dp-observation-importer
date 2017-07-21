package dimension

// OrderCache is cache the for order of dimensions in the input file
type OrderCache struct {
	orderStore OrderStore
}

// OrderStore represents the data store for dimension order
type OrderStore interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewOrderCache returns a new instance of the order cache that uses the given OrderStore.
func NewOrderCache(orderStore OrderStore) *OrderCache {
	return &OrderCache{
		orderStore:orderStore,
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (cache *OrderCache) GetOrder(instanceID string) ([]string, error) {

	// todo: implement an in memory cache for the store response.

	return cache.orderStore.GetOrder(instanceID)
}
