package event

// ObservationExtracted is the data that is output for each observation extracted.
type ObservationExtracted struct {
	Row        string `avro:"row"`
	InstanceID string `avro:"instance_id"`
}
