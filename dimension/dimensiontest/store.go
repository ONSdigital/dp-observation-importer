package dimensiontest

import (
	"errors"
)

// MockOrderStore provides a mock implementation of an order store.
type MockOrderStore struct {
	ReturnError bool
}

// GetOrder returns a mock response for the order of dimensions.
func (mos MockOrderStore) GetOrder(instanceID string) ([]string, error) {
	if mos.ReturnError {
		return []string{}, errors.New("Internal store error")
	}
	return []string{"V4_1", "data_marking", "time_codelist", "time"}, nil
}

// MockIDStore provides a mock implementation of an ID store.
type MockIDStore struct {
	ReturnError bool
}

// GetIDs returns a mock response for the ID's of dimensions.
func (mis MockIDStore) GetIDs(instanceID string) (map[string]string, error) {

	if mis.ReturnError {
		return nil, errors.New("Internal store error")
	}
	cache := make(map[string]string)
	cache["age_55"] = "123"

	return cache, nil
}
