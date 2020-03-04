package dimension

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"errors"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
)

const authorizationHeader = "Authorization"

// ErrParseAPIResponse used when the dataset API response fails to be parsed.
var ErrParseAPIResponse = errors.New("failed to parse dataset api response")

// ErrInstanceNotFound returned when the given instance ID is not found in the dataset API.
var ErrInstanceNotFound = errors.New("dataset api failed to find instance")

// ErrInternalError returned when an unrecognised internal error occurs in the dataset API.
var ErrInternalError = errors.New("internal error from the dataset api")

type csvHeaders struct {
	Headers []string `json:"headers"`
}

// NodeResults wraps dimension node objects for pagination
type NodeResults struct {
	Items []Dimension `json:"items"`
}

// Dimension which has been cached from the dataset api
type Dimension struct {
	DimensionName string `json:"dimension"`
	Option        string `json:"option"`
	NodeID        string `json:"node_id"`
}

// DatasetStore represents the storage of dimension data.
type DatasetStore struct {
	authToken        string
	datasetAPIURL    string
	datasetAPIClient DatasetClient
}

// DatasetClient represents the dataset client for dataset API
type DatasetClient interface {
	GetInstanceByBytes(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, instanceID string) ([]byte, error)
	GetInstanceDimensionsByBytes(ctx context.Context, userAuthToken, serviceAuthToken, instanceID string) (b []byte, err error)
}

// NewStore returns a new instance of a dimension store.
func NewStore(authToken, datasetAPIURL string, client DatasetClient) *DatasetStore {
	return &DatasetStore{
		authToken:        authToken,
		datasetAPIURL:    datasetAPIURL,
		datasetAPIClient: client,
	}
}

// GetOrder returns list of dimension names in the order they are stored in the input file.
func (store *DatasetStore) GetOrder(ctx context.Context, instanceID string) ([]string, error) {
	b, clientErr := store.datasetAPIClient.GetInstanceByBytes(ctx, "", store.authToken, "", instanceID)
	if err := checkResponse(clientErr); err != nil {
		return nil, err
	}

	var csv csvHeaders
	JSONErr := json.Unmarshal(b, &csv)
	if JSONErr != nil {
		return nil, JSONErr
	}
	return csv.Headers, nil
}

// GetIDs returns all dimensions for a given instanceID
func (store *DatasetStore) GetIDs(ctx context.Context, instanceID string) (map[string]string, error) {
	b, clientErr := store.datasetAPIClient.GetInstanceDimensionsByBytes(ctx, "", store.authToken, instanceID)
	if err := checkResponse(clientErr); err != nil {
		return nil, err
	}

	var dimensionResults NodeResults
	JSONErr := json.Unmarshal(b, &dimensionResults)
	if JSONErr != nil {
		return nil, JSONErr
	}
	cache := make(map[string]string)
	for _, dimension := range dimensionResults.Items {
		cache[fmt.Sprintf("%s_%s_%s", instanceID, dimension.DimensionName, dimension.Option)] = dimension.NodeID
	}

	return cache, nil
}

func checkResponse(err error) error {
	if err == nil {
		return nil
	}

	// Get status code from error
	switch err.(type) {
	case *dataset.ErrInvalidDatasetAPIResponse:
		code := err.(*dataset.ErrInvalidDatasetAPIResponse).Code()

		switch code {
		case http.StatusNotFound:
			return ErrInstanceNotFound
		case http.StatusInternalServerError:
			return ErrInternalError
		}
	}

	return err
}
