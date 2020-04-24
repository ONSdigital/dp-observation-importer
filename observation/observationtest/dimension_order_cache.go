package observationtest

import (
	"context"

	"github.com/ONSdigital/dp-observation-importer/observation"
)

var _ observation.DimensionHeaderCache = (*DimensionHeaderCache)(nil)

// DimensionHeaderCache mock
type DimensionHeaderCache struct {
	DimensionOrder []string
	Error          error
}

// GetOrder returns the stored mock response.
func (cache DimensionHeaderCache) GetOrder(ctx context.Context, instanceID string) ([]string, error) {
	return cache.DimensionOrder, cache.Error
}
