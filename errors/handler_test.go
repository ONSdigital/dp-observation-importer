package errors_test

import (
	"errors"
	errs "github.com/ONSdigital/dp-observation-importer/errors"
	"github.com/ONSdigital/go-ns/kafka/kafkatest"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("Given an error handler", t, func() {

		outputChannel := make(chan []byte, 1)
		mockMessageProducer := kafkatest.NewMessageProducer(outputChannel, nil, nil)

		errorHandler := errs.NewKafkaHandler(mockMessageProducer)

		Convey("When handle is called with a nil log data", func() {

			errorHandler.Handle("123", errors.New("broken"), nil)

			Convey("The error message is written (no panics occur)", func() {
			})
		})
	})
}
