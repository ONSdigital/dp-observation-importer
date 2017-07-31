package observation

import (
	"strings"
	"fmt"
)

// Mapper interprets a CSV line and returns an observation instance.
type Mapper struct {
	dimensionCache DimensionOrderCache
	dimensionNodeIdStore DimensionNodeIdStore
}

// DimensionOrderCache provides the an array of dimension names to define the order of dimensions
type DimensionOrderCache interface {
	GetOrder(instanceID string) ([]string, error)
}

type DimensionNodeIdStore interface {
	GetNodeIDs(instanceID string) (map[string]string, error)
}

// NewMapper returns a new Mapper instance
func NewMapper(dimensionOrderCache DimensionOrderCache, dimensionNodeIdStore DimensionNodeIdStore) *Mapper {
	return &Mapper{
		dimensionCache:dimensionOrderCache,
		dimensionNodeIdStore:dimensionNodeIdStore,
	}
}

// Map the given CSV row to an observation instance.
func (mapper *Mapper) Map(row string, instanceID string) (*Observation, error) {
	header, err := mapper.dimensionCache.GetOrder(instanceID)
	nodeIdCache, err := mapper.dimensionNodeIdStore.GetNodeIDs(instanceID)
	if err != nil {
		return nil, err
	}
	var dimensions []DimensionOption
	csvRow := strings.Split(row, ",")
    offset := 2 // skip observation value / data markings
	for i := offset; i < len(header); i+=2 {
		codeListName := header[i]
		codeListValue := csvRow[i]
		labelValue := csvRow[i+1]
		dimensionLookUp := instanceID + "_" + codeListValue + "_" + labelValue

		nodeID, ok := nodeIdCache[dimensionLookUp]
		if ! ok {
			return nil, fmt.Errorf("No nodeId found for %s", dimensionLookUp)
		}
		dimensions = append(dimensions,
			DimensionOption{DimensionName:instanceID + "_" + codeListValue, NodeID:nodeID, NodeAlias:codeListName})

	}

    o := Observation{Row:row, InstanceID:instanceID, DimensionOptions:dimensions}
	return &o, nil
}
