package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/dp-observation-importer/service"
	"github.com/ONSdigital/log.go/v2/log"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

const serviceName = "dp-observation-importer"

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(ctx, "fatal runtime error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "failed to retrieve configuration", err)
		return err
	}

	// Sensitive fields are omitted from config.String()
	log.Info(ctx, "loaded config", log.Data{"config": cfg})

	// External services and their initialization state
	serviceList := initialise.NewServiceList(&initialise.Init{})

	if err := service.Run(ctx, cfg, serviceList, signals, BuildTime, GitCommit, Version); err != nil {
		log.Error(ctx, "application unexpectedly failed", err)
		return err
	}

	return nil
}
