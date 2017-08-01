package csv

import (
	"strconv"
	"strings"
)

// Header represent the CSV header row of an import file.
type Header struct {
	row []string
}

// NewHeader returns a new header instance for the given header row string.
func NewHeader(header []string) *Header {
	return &Header{
		row: header,
	}
}

// DimensionOffset returns the number of columns before the start of the dimension columns.
func (header Header) DimensionOffset() (int, error) {
	metaData := strings.Split(header.row[0], "_")
	dimensionColumnOffset, err := strconv.Atoi(metaData[1])
	if err != nil {
		return 0, err
	}

	return dimensionColumnOffset + 1, nil
}
