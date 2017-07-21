package errors

import "github.com/ONSdigital/go-ns/log"

type Handler interface {
	Handle(err error, data log.Data)
}

type KafkaHandler struct {
	//messageProducer MessageProducer
}
