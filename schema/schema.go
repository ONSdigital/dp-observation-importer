package schema

import (
	"github.com/ONSdigital/go-ns/avro"
)

var observationExtractedEvent = `{
  "type": "record",
  "name": "observation-extracted",
  "fields": [
    {"name": "instance_id", "type": "string"},
    {"name": "row", "type": "string"}
  ]
}`

// ObservationExtractedEvent is the Avro schema for each observation extracted.
var ObservationExtractedEvent *avro.Schema = &avro.Schema{
	Definition: observationExtractedEvent,
}
