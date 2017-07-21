package eventtest

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/dp-observation-importer/errors"
)

var _ errors.Handler = (*ErrorHandler)(nil)

type ErrorHandler struct {

}

func (handler ErrorHandler) Handle(err error, data log.Data) {

}



