package dimension

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/headers"
	"github.com/ONSdigital/dp-observation-importer/dimension/dimensiontest"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

const authToken = "coffee"

func TestStore_GetOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)

	Convey("Given a valid instanceId", t, func() {

		Convey("When the client returns an instances state", func() {
			data := csvHeaders{
				Headers: []string{"V4_1", "Data_Marking", "Time_codelist"},
			}
			b, err := json.Marshal(data)
			if err != nil {
				t.Errorf("unable to json marshal test data: %v", data)
			}
			mockDatasetClient.EXPECT().GetInstanceBytes(ctx, "", authToken, "", "1", headers.IfMatchAnyETag).Return(b, "", nil)

			Convey("Then the CSV headers are returned", func() {
				dataStore := &DatasetStore{
					authToken:        authToken,
					datasetAPIURL:    "http://localhost:80",
					datasetAPIClient: mockDatasetClient,
				}

				headers, err := dataStore.GetOrder(context.Background(), "1")
				So(err, ShouldBeNil)
				So(headers, ShouldContain, "V4_1")
				So(headers, ShouldContain, "Data_Marking")
				So(headers, ShouldContain, "Time_codelist")
			})
		})
	})
}

func TestStore_GetOrderReturnAnError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("Given an invalid URL", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{StatusCode: http.StatusInternalServerError}, "someUri",
		)

		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceBytes(ctx, "", authToken, "", "1", headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        authToken,
			datasetAPIURL:    "http://unknown-url:288100",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns an internal error", func() {
			Convey("The CSV headers contains nothing and an error is returned", func() {
				headers, err := dataStore.GetOrder(context.Background(), "1")
				So(headers, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInternalError)
			})
		})
	})

	Convey("Given the instance does not exist", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			}, "someUri",
		)
		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceBytes(ctx, "", authToken, "", "1", headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        authToken,
			datasetAPIURL:    "http://localhost:80",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns a not found error", func() {
			Convey("The CSV headers contains nothing and an error is returned", func() {
				headers, err := dataStore.GetOrder(context.Background(), "1")
				So(headers, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInstanceNotFound)
			})
		})
	})

	Convey("Given the request is unauthorised", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{StatusCode: http.StatusUnauthorized}, "someUri",
		)
		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceBytes(ctx, "", "", "", "1", headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        "",
			datasetAPIURL:    "http://localhost:80",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns a not found error", func() {
			Convey("The CSV headers contains nothing and and the original error is returned", func() {
				headers, err := dataStore.GetOrder(context.Background(), "1")
				So(headers, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errDatasetAPICodeServerError)
			})
		})
	})
}

func TestIDCache_GetIDsReturnError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	// Setup mocked dataset client
	mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)

	dataStore := &DatasetStore{
		authToken:        authToken,
		datasetAPIURL:    "http://localhost:80",
		datasetAPIClient: mockDatasetClient,
	}

	Convey("Given a valid instance id", t, func() {
		data := &NodeResults{
			Items: []Dimension{
				{
					DimensionName: "year",
					Option:        "1997",
					NodeID:        "123",
				},
			},
		}

		b, err := json.Marshal(data)
		if err != nil {
			t.Errorf("unable to json marshal test data: %v", data)
		}
		mockDatasetClient.EXPECT().GetInstanceDimensionsBytes(ctx, authToken, "1", &dataset.QueryParams{}, headers.IfMatchAnyETag).Return(b, "", nil)

		Convey("When the client api is called ", func() {
			Convey("A list of dimensions are returned", func() {
				cache, err := dataStore.GetIDs(context.Background(), "1")
				So(err, ShouldBeNil)
				So(cache["1_year_1997"], ShouldEqual, "123")
			})
		})
	})
}

func TestStore_GetIDsReturnAnError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	Convey("Given an invalid URL", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{StatusCode: http.StatusInternalServerError}, "someUri",
		)

		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceDimensionsBytes(ctx, authToken, "1", &dataset.QueryParams{}, headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        authToken,
			datasetAPIURL:    "http://unknown-url:288100",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns an internal error", func() {
			Convey("The cache is empty and an error is returned", func() {
				cache, err := dataStore.GetIDs(context.Background(), "1")
				So(cache, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInternalError)
			})
		})
	})

	Convey("Given the instance does not exist", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte{})),
			}, "someUri",
		)

		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceDimensionsBytes(ctx, authToken, "1", &dataset.QueryParams{}, headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        authToken,
			datasetAPIURL:    "http://localhost:80",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns a not found error", func() {
			Convey("The cache is empty and an error is returned", func() {
				cache, err := dataStore.GetIDs(context.Background(), "1")
				So(cache, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInstanceNotFound)
			})
		})
	})

	Convey("Given the request is unauthorised", t, func() {
		errDatasetAPICodeServerError := dataset.NewDatasetAPIResponse(
			&http.Response{StatusCode: http.StatusUnauthorized}, "someUri",
		)
		// Setup mocked dataset client
		mockDatasetClient := dimensiontest.NewMockDatasetClient(mockCtrl)
		mockDatasetClient.EXPECT().GetInstanceDimensionsBytes(ctx, "", "1", &dataset.QueryParams{}, headers.IfMatchAnyETag).Return(nil, "", errDatasetAPICodeServerError)

		dataStore := &DatasetStore{
			authToken:        "",
			datasetAPIURL:    "http://localhost:80",
			datasetAPIClient: mockDatasetClient,
		}

		Convey("When the client returns a not found error", func() {
			Convey("The cache is empty and and the original error is returned", func() {
				cache, err := dataStore.GetIDs(context.Background(), "1")
				So(cache, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errDatasetAPICodeServerError)
			})
		})
	})
}
