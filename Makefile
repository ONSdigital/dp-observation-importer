SHELL=bash

BUILD=build
BUILD_ARCH=$(BUILD)/$(GOOS)-$(GOARCH)
BIN_DIR?=.

export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)

export GRAPH_DRIVER_TYPE?=neptune
export GRAPH_ADDR?=ws://localhost:8182/gremlin

build:
	@mkdir -p $(BUILD_ARCH)/$(BIN_DIR)
	go build -o $(BUILD_ARCH)/$(BIN_DIR)/dp-observation-importer cmd/dp-observation-importer/main.go
debug: build
	GRAPH_DRIVER_TYPE=$(GRAPH_DRIVER_TYPE) GRAPH_ADDR=$(GRAPH_ADDR) HUMAN_LOG=1 go run cmd/dp-observation-importer/main.go
test:
	go test -cover $(shell go list ./... | grep -v /vendor/)
.PHONY: build debug test
