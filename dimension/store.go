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
func (store *Store) GetIDs(instanceID string) (IDs, error) {

	//todo call import API for dimension database ID's

	return nil, nil
}


// IDs a map from the dimension name to its options
type IDs map[string]OptionIDs

// OptionIDs represents a single dimension option, including its database ID.
type OptionIDs map[string]string
