package observation

import (
	"encoding/csv"
	"strings"

	inputcsv "github.com/ONSdigital/dp-observation-importer/csv"
	"github.com/ONSdigital/dp-observation-importer/models"
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
func (mapper *Mapper) Map(row string, rowIndex int64, instanceID string) (*models.Observation, error) {

	headerRow, err := mapper.dimensionCache.GetOrder(instanceID)
	if err != nil {
		return nil, err
	}

	header := inputcsv.NewHeader(headerRow)
	dimensionOffset, err := header.DimensionOffset()
	if err != nil {
		return nil, err
	}

	var dimensions []*models.DimensionOption
	csv := csv.NewReader(strings.NewReader(row))
	csvRow, err := csv.Read()
	if err != nil {
		return nil, err
	}

	for i := dimensionOffset; i < len(headerRow); i += 2 {

		dimensionOption := csvRow[i]

		// if there is no code provided, use the label
		if len(dimensionOption) == 0 {
			dimensionOption = csvRow[i+1]
		}

		dimensionName := headerRow[i+1]

		dimensions = append(dimensions,
			&models.DimensionOption{DimensionName: dimensionName, Name: dimensionOption})
	}

	o := models.Observation{Row: row, RowIndex: rowIndex, InstanceID: instanceID, DimensionOptions: dimensions}

	return &o, nil
}
