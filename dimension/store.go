package dimension

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ImportAPIClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type csvHeaders struct {
	Headers []string `json:"headers"`
}

type Dimension struct {
	DimensionName string `json:"dimension_name"`
	Value         string `json:"value"`
	NodeId        string `json:"node_id"`
}

// Store represents the storage of dimension data.
type Store struct {
	importAPIURL    string
	importAPIClient ImportAPIClient
}

// NewStore returns a new instance of a dimension store.
func NewStore(importAPIURL string, client ImportAPIClient) *Store {
	return &Store{
		importAPIURL:    importAPIURL,
		importAPIClient: client,
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (store *Store) GetOrder(instanceID string) ([]string, error) {
	url := store.importAPIURL + "/instances/" + instanceID
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}
	bytes, err := store.processRequest(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to http body into bytes")
	}
	var csv csvHeaders
	JSONErr := json.Unmarshal(bytes, &csv)
	if JSONErr != nil {
		return nil, JSONErr
	}
	return csv.Headers, nil
}

// GetIDs returns all dimensions for a given instanceID
func (store *Store) GetIDs(instanceID string) (IDs, error) {

	url := store.importAPIURL + "/instances/" + instanceID + "/dimensions"
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}
	bytes, err := store.processRequest(request)
	if err != nil {
		return nil, fmt.Errorf("Failed to read http body into bytes")
	}

	var dimensions []Dimension
	JSONErr := json.Unmarshal(bytes, &dimensions)
	if JSONErr != nil {
		return nil, JSONErr
	}
	cache := make(map[string]string)
	for _, dimension := range dimensions {
		cache[dimension.DimensionName] = dimension.NodeId
	}
	return cache, nil
}

func (store *Store) processRequest(r *http.Request) ([]byte, error) {
	response, responseError := store.importAPIClient.Do(r)
	if responseError != nil {
		return nil, responseError
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read http body into bytes")
	}
	return bytes, nil
}

// IDs a map from the dimension name to its options
type IDs map[string]string

// OptionIDs represents a single dimension option, including its database ID.
type OptionIDs map[string]string
