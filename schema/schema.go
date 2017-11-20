package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var observationExtractedEvent = `{
  "type": "record",
  "name": "observation-extracted",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "row", "type": "string"},
    {"name": "row_index", "type": "long"}
  ]
}`

// ObservationExtractedEvent is the Avro schema for each observation extracted.
var ObservationExtractedEvent = &avro.Schema{
	Definition: observationExtractedEvent,
}

var observationsInsertedEvent = `{
  "type": "record",
  "name": "import-observations-inserted",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "observations_inserted", "type": "int"}
  ]
}`

// ObservationsInsertedEvent is the Avro schema for each observation batch inserted.
var ObservationsInsertedEvent = &avro.Schema{
	Definition: observationsInsertedEvent,
}
