package observationtest

import (
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/observation"
)

var _ observation.DimensionIDCache = (*DimensionIDCache)(nil)

// DimensionIDCache mock
type DimensionIDCache struct {
	InstanceID string
	IDs dimension.IDs
	Error error
}

// GetIDs captures the given parameters and returns the stored mock response.
func (cache DimensionIDCache) GetIDs(instanceID string) (dimension.IDs, error) {
	cache.InstanceID = instanceID
	return cache.IDs, cache.Error
}