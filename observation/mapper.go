package observation

import (
	"encoding/csv"
	"errors"
	"strings"

	inputcsv "github.com/ONSdigital/dp-observation-importer/csv"
	"github.com/ONSdigital/go-ns/log"
)

// Mapper interprets a CSV line and returns an observation instance.
type Mapper struct {
	dimensionCache DimensionHeaderCache
}

// DimensionHeaderCache provides the an array of dimension names to define the order of dimensions (v4 format)
type DimensionHeaderCache interface {
	GetOrder(instanceID string) ([]string, error)
}

// NewMapper returns a new Mapper instance
func NewMapper(dimensionOrderCache DimensionHeaderCache) *Mapper {
	return &Mapper{
		dimensionCache: dimensionOrderCache,
	}
}

// Map the given CSV row to an observation instance.
func (mapper *Mapper) Map(row string, instanceID string) (*Observation, error) {

	headerRow, err := mapper.dimensionCache.GetOrder(instanceID)
	if err != nil {
		return nil, err
	}
	if len(headerRow) < 1 {
		return nil, errors.New("No header available with instance.")
	}

	header := inputcsv.NewHeader(headerRow)

	dimensionOffset, err := header.DimensionOffset()
	if err != nil {
		log.ErrorC("Dimension offset error", err, nil)
		return nil, err
	}

	var dimensions []*DimensionOption
	csv := csv.NewReader(strings.NewReader(row))
	csvRow, err := csv.Read()
	if err != nil {
		return nil, err
	}

	// if we want to support time being in any column we need to look at the header to see what column time is in.
	// TODO review whether this is safe to assume. If there are a lot of rows it could be a lot of CPU to determine
	// this for every row.
	timeDimensionOffset := dimensionOffset // Assume time is always first

	for i := dimensionOffset; i < len(headerRow); i += 2 {

		var dimensionName, dimensionOption string

		if i == timeDimensionOffset {
			dimensionOption = csvRow[i+1] // code list value
		} else {

			dimensionOption = csvRow[i] // code list value

			// if there is no code provided, use the label
			if len(dimensionOption) == 0 {
				dimensionOption = csvRow[i+1] // dimension option value
			}
		}

		dimensionName = headerRow[i+1]

		dimensions = append(dimensions,
			&DimensionOption{DimensionName: dimensionName, Name: dimensionOption})
	}

	o := Observation{Row: row, InstanceID: instanceID, DimensionOptions: dimensions}

	return &o, nil
}
