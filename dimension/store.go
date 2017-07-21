package dimension

// Store represents the storage of dimension data.
type Store struct {
	importAPIURL string
}

// NewStore returns a new instance of a dimension store.
func NewStore(importAPIURL string) *Store {
	return &Store{
		importAPIURL:importAPIURL,
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (store *Store) GetOrder(instanceID string) ([]string, error) {

	// todo: call import API for dimension order array (inferred from CSV header)

	return nil, nil
}

// GetIDs returns all dimensions for a given instanceID
func (store *Store) GetIDs(instanceID string) ([]*Dimension, error) {

	//todo call import API for dimension database ID's

	return nil, nil
}

// Dimension represents a single dimension with all of its values.
type Dimension struct {
	Name   string
	Values []*Option
}

// Option represents a single dimension option, including its database ID.
type Option struct {
	Name string
	ID   int
}
