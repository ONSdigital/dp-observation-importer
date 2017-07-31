package dimensiontest

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MockImportApi struct {
	FailRequest bool
	Data        string
}

func (i MockImportApi) Do(req *http.Request) (*http.Response, error) {
	if i.FailRequest {
		return nil, fmt.Errorf("Failed to process the request")
	}
	body := strings.NewReader(i.Data)
	response := http.Response{StatusCode: http.StatusOK, Body: IOReadCloser{body}}
	return &response, nil
}

type IOReadCloser struct {
	io.Reader
}

func (IOReadCloser) Close() error { return nil }
