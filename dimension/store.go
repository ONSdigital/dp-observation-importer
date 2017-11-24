package dimension

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"errors"
)

// ErrParseAPIResponse used when the dataset API response fails to be parsed.
var ErrParseAPIResponse = errors.New("failed to parse dataset api response")

// ErrInstanceNotFound returned when the given instance ID is not found in the dataset API.
var ErrInstanceNotFound = errors.New("dataset api failed to find instance")

// ErrInternalError returned when an unrecognised internal error occurs in the dataset API.
var ErrInternalError = errors.New("internal error from the dataset api")

// DatasetAPIClient an interface used to access the dataset api
type DatasetAPIClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type csvHeaders struct {
	Headers []string `json:"headers"`
}

// DimensionNodeResults wraps dimension node objects for pagination
type DimensionNodeResults struct {
	Items []Dimension `json:"items"`
}

// Dimension which has been cached from the dataset api
type Dimension struct {
	DimensionName string `json:"dimension_id"`
	Option        string `json:"option"`
	NodeID        string `json:"node_id"`
}

// Store represents the storage of dimension data.
type Store struct {
	datasetAPIURL    string
	datasetAPIToken  string
	datasetAPIClient DatasetAPIClient
}

// NewStore returns a new instance of a dimension store.
func NewStore(datasetAPIURL, datasetAPIAuthToken string, client DatasetAPIClient) *Store {
	return &Store{
		datasetAPIURL:    datasetAPIURL,
		datasetAPIToken:  datasetAPIAuthToken,
		datasetAPIClient: client,
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (store *Store) GetOrder(instanceID string) ([]string, error) {
	url := store.datasetAPIURL + "/instances/" + instanceID
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	request.Header.Set("internal-token", store.datasetAPIToken)

	bytes, err := store.processRequest(request, instanceID)
	if err != nil {
		return nil, err
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

	url := store.datasetAPIURL + "/instances/" + instanceID + "/dimensions"
	request, requestErr := http.NewRequest("GET", url, nil)
	if requestErr != nil {
		return nil, requestErr
	}

	request.Header.Set("internal-token", store.datasetAPIToken)

	bytes, err := store.processRequest(request, instanceID)
	if err != nil {
		return nil, err
	}

	var dimensionResults DimensionNodeResults
	JSONErr := json.Unmarshal(bytes, &dimensionResults)
	if JSONErr != nil {
		return nil, JSONErr
	}
	cache := make(map[string]string)
	for _, dimension := range dimensionResults.Items {
		cache[fmt.Sprintf("%s_%s_%s", instanceID, dimension.DimensionName, dimension.Option)] = dimension.NodeID
	}
	return cache, nil
}

func (store *Store) processRequest(r *http.Request, instanceID string) ([]byte, error) {
	response, responseError := store.datasetAPIClient.Do(r)
	if responseError != nil {
		return nil, responseError
	}
	switch response.StatusCode {
	case http.StatusNotFound:
		return nil, ErrInstanceNotFound
	case http.StatusInternalServerError:
		return nil, ErrInternalError
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, ErrParseAPIResponse
	}
	return bytes, nil
}
