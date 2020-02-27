SHELL=bash

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

export SERVICE_AUTH_TOKEN=1sGnFikGSPjg3WKjFvdrvsdQNce7Jh7hZjDjzoh9YBsLG4kFlpLA2pjRbAN26MCM

build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/dp-observation-importer cmd/dp-observation-importer/main.go

debug: build
	GRAPH_DRIVER_TYPE="neo4j" GRAPH_ADDR="bolt://localhost:7687" HUMAN_LOG=1 go run $(LDFLAGS) cmd/dp-observation-importer/main.go

test:
	go test -cover -race ./...

.PHONY: build debug test
