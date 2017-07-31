package observation

import (
	"strings"
)

// Mapper interprets a CSV line and returns an observation instance.
type Mapper struct {
	dimensionCache DimensionOrderCache
}

// DimensionOrderCache provides the an array of dimension names to define the order of dimensions
type DimensionOrderCache interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewMapper returns a new Mapper instance
func NewMapper(dimensionOrderCache DimensionOrderCache) *Mapper {
	return &Mapper{
		dimensionCache: dimensionOrderCache,
	}
}

// Map the given CSV row to an observation instance.
func (mapper *Mapper) Map(row string, instanceID string) (*Observation, error) {
	header, err := mapper.dimensionCache.GetOrder(instanceID)
	if err != nil {
		return nil, err
	}
	var dimensions []DimensionOption
	csvRow := strings.Split(row, ",")
	offset := 2 // skip observation value / data markings
	for i := offset; i < len(header); i += 2 {
		dimensionName := header[i+1] // Sex
		dimensionOption := csvRow[i] // codelist value

		if len(dimensionOption) == 0 {
			dimensionOption = csvRow[i+1] // dimension option value
		}

		dimensions = append(dimensions,
			DimensionOption{DimensionName: dimensionName, Name: dimensionOption})
	}

	o := Observation{Row: row, InstanceID: instanceID, DimensionOptions: dimensions}
	return &o, nil
}
