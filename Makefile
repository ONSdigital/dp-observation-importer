build:
	go build -o build/dp-observation-importer
debug: build default-env
	HUMAN_LOG=1 ./build/dp-observation-importer
test:
	go test ./...
.PHONY: build debug default-env
