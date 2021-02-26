package service

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

type Health struct {
	hc         *healthcheck.HealthCheck
	httpServer *dphttp.Server
	context    context.Context
	isRunning  bool
}

func StartHealthChecker(s *Service, sc *ServiceContainer, serviceList *initialise.ExternalServiceList) (*Health, error) {

	hc, err := serviceList.GetHealthCheck(s.cfg, s.BuildTime, s.GitCommit, s.Version)
	if err != nil {
		log.Event(s.ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}

	// Add dataset API and graph checks
	if err := registerCheckers(s.ctx, &hc, sc.SyncConsumerGroup, sc.ObservationsImportedProducer, sc.ObservationsImportedErrProducer, *sc.DatasetClient, sc.GraphDB); err != nil {
		return nil, err
	}

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(hc.Handler)
	httpServer := dphttp.NewServer(s.cfg.BindAddr, router)

	// Disable auto handling of os signals by the HTTP server. This is handled
	// in the service so we can gracefully shutdown resources other than just
	// the HTTP server.
	httpServer.HandleOSSignals = false

	// a channel to signal a server error
	errorChannel := make(chan error)

	go func() {
		log.Event(s.ctx, "starting http server", log.INFO, log.Data{"bind_addr": s.cfg.BindAddr})
		if err = httpServer.ListenAndServe(); err != nil {
			errorChannel <- err
		}
	}()

	hc.Start(s.ctx)

	s.errorChan = errorChannel

	health := &Health{
		hc:         &hc,
		httpServer: httpServer,
		context:    s.ctx,
		isRunning:  serviceList.HealthCheck,
	}

	return health, nil
}

func (h *Health) Stop(shutdownContext context.Context) bool {

	log.Event(h.context, "shutting down health checker", log.INFO)

	if h.isRunning {
		h.hc.Stop()
	}

	var hasShutdownError = false
	err := h.httpServer.Shutdown(shutdownContext)
	if err != nil {
		log.Event(h.context, "failed to shutdown http server", log.ERROR, log.Error(err))
		hasShutdownError = true
	}

	return hasShutdownError
}
