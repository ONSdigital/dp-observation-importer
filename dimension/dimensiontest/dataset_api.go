package dimensiontest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MockDatasetAPI provides mock functionality for the dataset API.
type MockDatasetAPI struct {
	FailRequest bool
	Data        string
}

// Do returns a mock HTTP response for the given request.
func (i MockDatasetAPI) Do(req *http.Request) (*http.Response, error) {
	if i.FailRequest {
		return nil, fmt.Errorf("Failed to process the request")
	}
	body := strings.NewReader(i.Data)
	response := http.Response{StatusCode: http.StatusOK, Body: iOReadCloser{body}}
	return &response, nil
}

type iOReadCloser struct {
	io.Reader
}

func (iOReadCloser) Close() error { return nil }
