package observation

// Observation represents a single observation value and its associated data.
type Observation struct {
	Row              string
	InstanceID       string
	DimensionOptions []DimensionOption
}

// DimensionOption represents the a single dimension option for an observation.
type DimensionOption struct {
	DimensionName string
	Name          string
}
