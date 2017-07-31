package eventtest

import (
	"github.com/ONSdigital/dp-observation-importer/errors"
	"github.com/ONSdigital/go-ns/log"
)

var _ errors.Handler = (*ErrorHandler)(nil)

// ErrorHandler is a mock error handler.
type ErrorHandler struct {
}

// Handle the error
func (handler ErrorHandler) Handle(err error, data log.Data) {
}
