package dimension

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"net/http"
	"strings"
	"io"
	"fmt"
)

func TestStore_GetOrder(t *testing.T) {

	data := "{\"headers\": [\"V4_1\",\"Data_Marking\",\"Time_codelist\"]}"
	dataStore := NewStore("http://localhost:288100", MockImportApi{Data:data})

	Convey("Given a valid instanceId", t, func() {

		Convey("When the client returns an instances state", func() {
			Convey("The CSV headers are returned", func() {
				headers, error := dataStore.GetOrder("1")
				So(error, ShouldBeNil)
				So(headers, ShouldContain, "V4_1")
				So(headers, ShouldContain, "Data_Marking")
				So(headers, ShouldContain, "Time_codelist")
			})
		})
	})
}

func TestStore_GetOrderReturnAnError(t *testing.T) {

	dataStore := NewStore("http://unknown-url:288100", MockImportApi{FailRequest:true})

	Convey("Given a invalid URL", t, func() {

		Convey("When the client returns an error", func() {
			Convey("The CSV headers contains nothing and an error is returned", func() {
				headers, error := dataStore.GetOrder("1")
				So(headers, ShouldBeNil)
				So(error, ShouldNotBeNil)
			})
		})
	})
}

func TestIDCache_GetIDs(t *testing.T) {
	data := "[{ \"dimension_name\": \"6_Year_1997\",\"value\": \"1997\",\"node_id\": \"123\"}]"
	dataStore := NewStore("http://localhost:288100", MockImportApi{Data:data})
	Convey("Given a valid instance id", t, func() {
		Convey("When the client api is called ", func() {
			Convey("A list of dimensions are returned", func() {
				dimensions, error := dataStore.GetIDs("1")
				So(error, ShouldBeNil)
				So(dimensions["_1997"], ShouldEqual, "123")
			})
		})
	})
}

type MockImportApi struct {
	FailRequest bool
	Data string

}

func (i MockImportApi) Do(req *http.Request) (*http.Response, error) {
	if i.FailRequest {
		return nil, fmt.Errorf("Failed to process the request")
	}
	body := strings.NewReader(i.Data)
	response := http.Response{StatusCode: http.StatusOK, Body:IOReadCloser{body}}
	return &response, nil
}

type IOReadCloser struct {
	io.Reader
}

func (IOReadCloser) Close() error { return nil }