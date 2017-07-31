package observationtest

import (
	"github.com/ONSdigital/dp-observation-importer/observation"
)

var _ observation.DimensionHeaderCache = (*DimensionHeaderCache)(nil)

// DimensionHeaderCache mock
type DimensionHeaderCache struct {
	DimensionOrder []string
	Error          error
}

// GetHeader returns the stored mock response.
func (cache DimensionHeaderCache) GetOrder(instanceID string) ([]string, error) {
	return cache.DimensionOrder, cache.Error
}
