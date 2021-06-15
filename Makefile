SHELL=bash

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

export GRAPH_DRIVER_TYPE?=neo4j
export GRAPH_ADDR?=bolt://localhost:7687

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -json -m all | nancy sleuth

.PHONY: build
build:
	@mkdir -p $(BUILD)/$(BIN_DIR)
	go build $(LDFLAGS) -o $(BUILD)/$(BIN_DIR)/dp-observation-importer cmd/dp-observation-importer/main.go

.PHONY: debug
debug: build
	HUMAN_LOG=1 go run -race $(LDFLAGS) cmd/dp-observation-importer/main.go

.PHONY: lint
lint:
	exit

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: test-component
test-component:
	go test -race -cover -coverpkg=github.com/ONSdigital/dp-observation-importer/... -component
