package dimension

import (
	"encoding/json"
	"fmt"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/errors"
	"io/ioutil"
	"net/http"
)

// ImportAPIClient an interface used to access the import api
type ImportAPIClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type csvHeaders struct {
	Headers []string `json:"headers"`
}

// Dimension which has been cached from the import api
type Dimension struct {
	DimensionName string `json:"dimension_id"`
	Value         string `json:"value"`
	NodeID        string `json:"node_id"`
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
	bytes, err := store.processRequest(request, instanceID)
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
func (store *Store) GetIDs(instanceID string) (map[string]string, error) {

	url := store.importAPIURL + "/instances/" + instanceID + "/dimensions"
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	bytes, err := store.processRequest(request, instanceID)
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
		cache[(dimension.DimensionName + "_" + dimension.Value)] = dimension.NodeID
	}
	return cache, nil
}

func (store *Store) processRequest(r *http.Request, instanceID string) ([]byte, error) {
	response, responseError := store.importAPIClient.Do(r)
	if responseError != nil {
		return nil, responseError
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, errors.New("Failed to find instanceId : " + instanceID)
	case http.StatusInternalServerError:
		return nil, errors.New("Internal error from the import api")
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read http body into bytes")
	}
	return bytes, nil
}
