package service

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-graph/v2/graph"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-reporter-client/reporter"
	"github.com/ONSdigital/log.go/log"
)

type ServiceContainer struct {
	SyncConsumerGroup               kafka.IConsumerGroup
	ObservationsImportedProducer    kafka.IProducer
	ObservationsImportedErrProducer kafka.IProducer
	ErrorReporter                   reporter.ImportErrorReporter
	GraphDB                         *graph.DB
	DatasetClient *dataset.Client
}

func BuildServiceContainer(ctx context.Context, cfg *config.Config, serviceList *initialise.ExternalServiceList) (*ServiceContainer, error) {
	container := &ServiceContainer{}

	// Get syncConsumerGroup Kafka Consumer
	syncConsumerGroup, err := serviceList.GetConsumer(ctx, cfg)
	if err != nil {
		log.Event(ctx, "could not obtain consumer group", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get observations inserted Kafka Producer
	observationsImportedProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ResultProducerTopic,
		initialise.ObservationsImported,
		cfg,
	)
	if err != nil {
		log.Event(ctx, "could not obtain observations inserted producer", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get observations inserted error Kafka Producer
	observationsImportedErrProducer, err := serviceList.GetProducer(
		ctx,
		cfg.Brokers,
		cfg.ErrorProducerTopic,
		initialise.ObservationsImportedErr,
		cfg,
	)
	if err != nil {
		log.Event(ctx, "could not obtain observations inserted error producer", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get Error reporter
	errorReporter, err := serviceList.GetImportErrorReporter(observationsImportedErrProducer, log.Namespace)
	if err != nil {
		log.Event(ctx, "error while attempting to create error reporter client", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get graphdb connection for observation store
	graphDB, err := serviceList.GetGraphDB(ctx)
	if err != nil {
		log.Event(ctx, "failed to instantiate graph observation store", log.FATAL, log.Error(err))
		return nil, err
	}

	container.SyncConsumerGroup = syncConsumerGroup
	container.ErrorReporter = errorReporter
	container.GraphDB = graphDB
	container.ObservationsImportedErrProducer = observationsImportedErrProducer
	container.ObservationsImportedProducer = observationsImportedProducer
	container.DatasetClient = dataset.NewAPIClient(cfg.DatasetAPIURL)

	return container, nil
}