package service

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v2"
	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/dimension"
	"github.com/ONSdigital/dp-observation-importer/event"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-observation-importer/observation"
	"github.com/ONSdigital/log.go/log"
	"github.com/pkg/errors"
	"os"
)

type Service struct {
	BuildTime string
	GitCommit string
	Version   string
	ctx       context.Context
	cfg       *config.Config
	errorChan chan error
}

func New(
	config *config.Config,
	context context.Context,
	BuildTime string,
	GitCommit string,
	Version string,
) *Service {
	return &Service{
		BuildTime: BuildTime,
		GitCommit: GitCommit,
		Version:   Version,
		ctx:       context,
		cfg:       config,
	}
}

func (s *Service) Run(serviceList *initialise.ExternalServiceList, signals chan os.Signal) error {

	log.Event(s.ctx, "starting observation importer", log.INFO)
	sc, err := BuildServiceContainer(s.ctx, s.cfg, serviceList)
	if err != nil {
		return err
	}

	var graphErrorConsumer *graph.ErrorConsumer
	if serviceList.Graph {
		graphErrorConsumer = graph.NewLoggingErrorConsumer(s.ctx, sc.GraphDB.Errors)
	}

	healthChecker, err := StartHealthChecker(s, sc, serviceList)
	if err != nil {
		return err
	}

	// objects to get dimension data - via the dataset API + cached locally in memory.
	dimensionStore := dimension.NewStore(s.cfg.ServiceAuthToken, s.cfg.DatasetAPIURL, sc.DatasetClient)
	dimensionOrderCache := dimension.NewOrderCache(dimensionStore, s.cfg.CacheTTL)
	dimensionIDCache := dimension.NewIDCache(dimensionStore, s.cfg.CacheTTL)

	// maps from CSV row to observation data.
	observationMapper := observation.NewMapper(dimensionOrderCache)

	// stores observations in the DB.
	observationStore := observation.NewStore(dimensionIDCache, sc.GraphDB, sc.ErrorReporter)

	// write import results to kafka topic.
	resultWriter := observation.NewResultWriter(sc.ObservationsImportedProducer)

	// handle a batch of events.
	batchHandler := event.NewBatchHandler(observationMapper, observationStore, resultWriter, sc.ErrorReporter)

	eventConsumer := event.NewConsumer()

	// Start listening for event messages.
	eventConsumer.Consume(sc.SyncConsumerGroup, s.cfg.BatchSize, batchHandler, s.cfg.BatchWaitTime, s.errorChan)

	sc.SyncConsumerGroup.Channels().LogErrors(s.ctx, "error received from kafka consumer, topic: "+s.cfg.ObservationConsumerTopic)
	sc.ObservationsImportedProducer.Channels().LogErrors(s.ctx, "error received from kafka producer, topic: "+s.cfg.ResultProducerTopic)
	sc.ObservationsImportedErrProducer.Channels().LogErrors(s.ctx, "error received from kafka producer, topic: "+s.cfg.ErrorProducerTopic)

	// block until a fatal error occurs
	select {
	case <-signals:
		log.Event(s.ctx, "os signal received", log.INFO)
	case err = <-s.errorChan:
		log.Event(s.ctx, "error received from http server, shutting down application", log.ERROR, log.Error(err))
	}

	return s.Shutdown(serviceList, healthChecker, sc, eventConsumer, graphErrorConsumer)
}

func (s *Service) Shutdown(serviceList *initialise.ExternalServiceList, healthChecker *Health, sc *ServiceContainer, eventConsumer *event.Consumer, graphErrorConsumer *graph.ErrorConsumer) error {
	log.Event(s.ctx, fmt.Sprintf("shutdown with timeout: %s", s.cfg.GracefulShutdownTimeout), log.INFO)
	shutdownContext, cancel := context.WithTimeout(context.Background(), s.cfg.GracefulShutdownTimeout)

	// track shutdown gracefully closes app
	var gracefulShutdown bool

	// Gracefully shutdown the application closing any open resources.
	go func() {
		defer cancel()

		hasShutdownError := healthChecker.Stop(shutdownContext)

		// If kafka consumer exists, stop listening to it. (Will close later)
		if serviceList.Consumer {
			log.Event(shutdownContext, "stopping kafka consumer listener", log.INFO)
			sc.SyncConsumerGroup.StopListeningToConsumer(shutdownContext)
			log.Event(shutdownContext, "stopped kafka consumer listener", log.INFO)
		}

		// If observation imported kafka producer exists, close it
		if serviceList.ObservationsImportedProducer {
			log.Event(shutdownContext, "closing observation imported kafka producer", log.INFO)
			sc.ObservationsImportedProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported kafka producer", log.INFO)
		}

		// If observation imported error kafka producer exists, close it
		if serviceList.ObservationsImportedErrProducer {
			log.Event(shutdownContext, "closing observation imported error kafka producer", log.INFO)
			sc.ObservationsImportedErrProducer.Close(shutdownContext)
			log.Event(shutdownContext, "closed observation imported error kafka producer", log.INFO)
		}

		// Attempting to close event consumer
		log.Event(shutdownContext, "closing event wrapper", log.INFO)
		eventConsumer.Close(shutdownContext)
		log.Event(shutdownContext, "closed event wrapper", log.INFO)

		// If kafka consumer exists, close it.
		if serviceList.Consumer {
			log.Event(shutdownContext, "closing kafka consumer", log.INFO)
			sc.SyncConsumerGroup.Close(shutdownContext)
			log.Event(shutdownContext, "closed kafka consumer", log.INFO)
		}

		if serviceList.Graph {
			log.Event(s.ctx, "closing graph db", log.INFO)
			err := sc.GraphDB.Close(s.ctx)
			if err != nil {
				log.Event(s.ctx, "failed to close graph db", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			err = graphErrorConsumer.Close(s.ctx)
			if err != nil {
				log.Event(s.ctx, "failed to close graph db error consumer", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
			log.Event(s.ctx, "closed graph db", log.INFO)
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Event(shutdownContext, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(shutdownContext, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the checkers for the provided clients to the healthcheck object
func registerCheckers(ctx context.Context, hc *healthcheck.HealthCheck,
	kafkaConsumer kafka.IConsumerGroup,
	observationsImportedProducer kafka.IProducer,
	observationsImportedErrProducer kafka.IProducer,
	datasetClient dataset.Client,
	graphDB *graph.DB) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Kafka Consumer", kafkaConsumer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka consumer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Import Producer", observationsImportedProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka import producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Kafka Error Producer", observationsImportedErrProducer.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for kafka error producer", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Dataset API", datasetClient.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for dataset client", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Graph DB", graphDB.Driver.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for graph db", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("error registering checkers for health check")
	}
	return nil
}
