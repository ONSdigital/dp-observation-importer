package observation

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/ONSdigital/dp-graph/v2/models"

	inputcsv "github.com/ONSdigital/dp-observation-importer/csv"
)

// Mapper interprets a CSV line and returns an observation instance.
type Mapper struct {
	dimensionCache DimensionHeaderCache
}

// DimensionHeaderCache provides the an array of dimension names to define the order of dimensions (v4 format)
type DimensionHeaderCache interface {
	GetOrder(ctx context.Context, instanceID string) ([]string, error)
}

// NewMapper returns a new Mapper instance
func NewMapper(dimensionOrderCache DimensionHeaderCache) *Mapper {
	return &Mapper{
		dimensionCache: dimensionOrderCache,
	}
}

// Map the given CSV row to an observation instance.
func (mapper *Mapper) Map(ctx context.Context, row string, rowIndex int64, instanceID string) (*models.Observation, error) {

	headerRow, err := mapper.dimensionCache.GetOrder(ctx, instanceID)
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

	// Populate DimensionOptions map assuming that the code always corresponds to the first column.
	// Empty options will return an error, although in this scenario, dimension importer should already have failed.
	for i := dimensionOffset; i < len(headerRow); i += 2 {
		dimensionOption := csvRow[i] // code list value
		if len(dimensionOption) == 0 {
			return nil, fmt.Errorf("dimension option was empty, it was expected in the first column. Csv row: %v", csvRow)
		}
		dimensionName := headerRow[i+1]
		dimensions = append(dimensions,
			&models.DimensionOption{DimensionName: dimensionName, Name: dimensionOption})
	}

	o := models.Observation{Row: row, RowIndex: rowIndex, InstanceID: instanceID, DimensionOptions: dimensions}

	return &o, nil
}
