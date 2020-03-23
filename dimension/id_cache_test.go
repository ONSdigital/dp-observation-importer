package dimension

import (
	"context"
	"testing"
	"time"

	"github.com/ONSdigital/dp-observation-importer/dimension/dimensiontest"
	"github.com/golang/mock/gomock"
	cache "github.com/patrickmn/go-cache"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDimensionMemoryCache_GetNodeIDs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	data := make(map[string]string)
	data["1_year_1997"] = "123"

	mockIDStore := dimensiontest.NewMockIDStore(mockCtrl)
	mockIDStore.EXPECT().GetIDs(ctx, "1").Return(data, nil)

	cache := &MemoryCache{
		store:       mockIDStore,
		memoryCache: cache.New(1*time.Minute, 15*time.Minute),
	}

	Convey("Given a valid instance Id", t, func() {
		Convey("When a request for a cached dimensions", func() {
			dimensions, err := cache.GetNodeIDs(context.Background(), "1")
			Convey("The dimensions are returned", func() {
				So(err, ShouldBeNil)
				So(dimensions["1_year_1997"], ShouldEqual, "123")
			})
		})
	})
}

func TestDimensionMemoryCache_GetNodeIDsReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ctx := gomock.Any()

	mockIDStore := dimensiontest.NewMockIDStore(mockCtrl)
	mockIDStore.EXPECT().GetIDs(ctx, "1").Return(nil, ErrInstanceNotFound)

	cache := &MemoryCache{
		store:       mockIDStore,
		memoryCache: cache.New(1*time.Minute, 15*time.Minute),
	}

	Convey("Given a invalid instance Id", t, func() {

		Convey("When a request for cached dimensions", func() {
			dimensions, err := cache.GetNodeIDs(context.Background(), "1")
			Convey("An error is returned", func() {
				So(dimensions, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, ErrInstanceNotFound)
			})
		})
	})
}
