.PHONY: test build run worker
test:
	go test ./...
	go vet ./...
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/admin ./cmd/admin
	go build -o bin/initdb ./cmd/initdb
run:
	go run ./cmd/server
worker:
	go run ./cmd/worker
