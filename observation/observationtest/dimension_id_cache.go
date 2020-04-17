package observationtest

import (
	"context"

	"github.com/ONSdigital/dp-observation-importer/observation"
)

var _ observation.DimensionIDCache = (*DimensionIDCache)(nil)

// DimensionIDCache mock
type DimensionIDCache struct {
	InstanceID string
	IDs        map[string]string
	Error      error
}

// GetNodeIDs captures the given parameters and returns the stored mock response.
func (cache *DimensionIDCache) GetNodeIDs(ctx context.Context, instanceID string) (map[string]string, error) {
	cache.InstanceID = instanceID
	return cache.IDs, cache.Error
}
