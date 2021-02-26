package main

import (
	"context"
	"github.com/ONSdigital/dp-observation-importer/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-observation-importer/config"
	"github.com/ONSdigital/dp-observation-importer/initialise"
	"github.com/ONSdigital/log.go/log"
)

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string
)

func main() {
	log.Namespace = "dp-observation-importer"
	ctx := context.Background()

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "failed to retrieve configuration", log.FATAL, log.Error(err))
		os.Exit(1)
	}
	// External services and their initialization state
	// var serviceList initialise.ExternalServiceList
	serviceList := initialise.NewServiceList(&initialise.Init{})

	// Sensitive fields are omitted from config.String()
	log.Event(ctx, "loaded config", log.INFO, log.Data{"config": cfg})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	svc := service.New(cfg, ctx, BuildTime, GitCommit, Version)

	if err := svc.Run(&serviceList, signals); err != nil {
		log.Event(ctx, "application unexpectedly failed", log.ERROR, log.Error(err))
		os.Exit(1)
	}

	os.Exit(0)
}