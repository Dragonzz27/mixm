.PHONY: test build run worker migrate
test:
	go test ./...
	go vet ./...
build:
	go build -o bin/server ./cmd/server
	go build -o bin/worker ./cmd/worker
	go build -o bin/admin ./cmd/admin
	go build -o bin/initdb ./cmd/initdb
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/migrate-sqlite ./cmd/migrate-sqlite
run:
	go run ./cmd/server
worker:
	go run ./cmd/worker
migrate:
	go run ./cmd/migrate
