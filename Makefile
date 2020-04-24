SHELL=bash

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

export GRAPH_DRIVER_TYPE?=neptune
export GRAPH_ADDR?=ws://localhost:8182/gremlin

build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/dp-observation-importer cmd/dp-observation-importer/main.go

debug: build
	GRAPH_DRIVER_TYPE=$(GRAPH_DRIVER_TYPE) GRAPH_ADDR=$(GRAPH_ADDR) HUMAN_LOG=1 go run $(LDFLAGS) cmd/dp-observation-importer/main.go
test:
	go test -cover -race ./...

.PHONY: build debug test
