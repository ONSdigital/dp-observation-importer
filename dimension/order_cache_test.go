package dimension

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dp-observation-importer/dimension/dimensiontest"
	"github.com/golang/mock/gomock"
	"github.com/patrickmn/go-cache"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHeaderCache_GetOrder(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	mockOrderStore := dimensiontest.NewMockOrderStore(mockCtrl)
	mockOrderStore.EXPECT().GetOrder(ctx, "1").Return([]string{"V4_1", "data_marking", "time_codelist"}, nil)

	cache := &HeaderCache{
		orderStore:  mockOrderStore,
		memoryCache: cache.New(1*time.Minute, 15*time.Minute),
	}

	Convey("Given a valid instanceId", t, func() {

		Convey("When a request for a cached CSV header is made", func() {
			headers, err := cache.GetOrder(context.Background(), "1")

			Convey("The CSV headers is returned", func() {
				So(err, ShouldBeNil)
				So(headers, ShouldContain, "V4_1")
				So(headers, ShouldContain, "data_marking")
				So(headers, ShouldContain, "time_codelist")
			})
		})
	})
}

func TestHeaderCache_GetOrder_ReturnErrors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	mockOrderStore := dimensiontest.NewMockOrderStore(mockCtrl)
	mockOrderStore.EXPECT().GetOrder(ctx, "1").Return(nil, ErrInstanceNotFound)

	cache := &HeaderCache{
		orderStore:  mockOrderStore,
		memoryCache: cache.New(1*time.Minute, 15*time.Minute),
	}

	Convey("Given an invalid instanceId", t, func() {
		Convey("When a request for the CSV header is made", func() {
			cache, err := cache.GetOrder(context.Background(), "1")
			Convey("An error is returned", func() {
				So(cache, ShouldBeEmpty)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInstanceNotFound)
			})
		})
	})
}
