package dimension

import (
	"github.com/patrickmn/go-cache"
	"time"
)

// HeaderCache is cache the for order of dimensions in the input file
type HeaderCache struct {
	orderStore  OrderStore
	memoryCache *cache.Cache
}

// OrderStore represents the data store for dimension order
type OrderStore interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewOrderCache returns a new instance of the order cache that uses the given OrderStore.
func NewOrderCache(orderStore OrderStore, cacheTTL time.Duration) *HeaderCache {
	return &HeaderCache{
		orderStore:  orderStore,
		memoryCache: cache.New(cacheTTL, 15*time.Minute),
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (hc *HeaderCache) GetOrder(instanceID string) ([]string, error) {
	item, ok := hc.memoryCache.Get(instanceID)
	if ok {
		return item.([]string), nil
	}
	newHeaders, err := hc.orderStore.GetOrder(instanceID)
	if err != nil {
		return nil, err
	}
	err = hc.memoryCache.Add(instanceID, newHeaders, cache.DefaultExpiration)
	if err != nil {
		return []string{}, err
	}
	return newHeaders, nil
}
