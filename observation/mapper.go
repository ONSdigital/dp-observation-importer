package observation

// Mapper interprets a CSV line and returns an observation instance.
type Mapper struct {
	dimensionOrderCache DimensionOrderCache
}

// DimensionOrderCache provides the an array of dimension names to define the order of dimensions
type DimensionOrderCache interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewMapper returns a new Mapper instance
func NewMapper(dimensionOrderCache DimensionOrderCache) *Mapper {
	return &Mapper{
		dimensionOrderCache:dimensionOrderCache,
	}
}

// Map the given CSV row to an observation instance.
func (mapper *Mapper) Map(row string, instanceID string) (*Observation, error) {
	// get the CSV header from the cache and store it against the instance ID

	// convert csv row to observation including dimension data and instanceID

	return nil, nil
}
