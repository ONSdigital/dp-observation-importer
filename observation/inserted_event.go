package observation

// InsertedEvent is the data that is output for each observation batch inserted.
type InsertedEvent struct {
	ObservationsInserted int32  `avro:"observations_inserted"`
	InstanceID           string `avro:"instance_id"`
}
