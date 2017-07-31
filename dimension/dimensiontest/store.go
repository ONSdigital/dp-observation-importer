package dimensiontest

import (
   "errors"
)
type MockOrderStore struct {
	ReturnError bool
}

func (mos MockOrderStore) GetOrder(instanceID string) ([]string, error) {
	if mos.ReturnError {
		return []string{}, errors.New("Internal store error")
	}
	return []string{"V4_1", "data_marking", "time_codelist", "time"}, nil
}

type MockIDStore struct {
	ReturnError bool
}

func (mis MockIDStore) GetIDs(instanceID string) (map[string]string, error) {

	if mis.ReturnError {
		return nil, errors.New("Internal store error")
	}
	cache := make(map[string]string)
	cache["age_55"] = "123"

	return cache, nil
}


