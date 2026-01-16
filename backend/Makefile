.PHONY: build run test clean docker-build docker-up docker-down proto

# Variables
BINARY_NAME=aether-server
PROTO_DIR=proto
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) ./cmd/server

# Run the server
run: build
	./$(BINARY_NAME) -config config.yaml

# Generate protobuf files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

# Run tests
test:
	go test -v ./...

# Run tests with race detection
test-race:
	go test -race -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run integration tests
test-integration:
	go test -tags=integration -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	go clean -cache

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run linter
lint:
	golangci-lint run

# Security scan
sec:
	gosec ./...

# Docker commands
docker-build:
	docker build -f docker/Dockerfile -t aether-server .

docker-up:
	cd docker && docker-compose up -d

docker-down:
	cd docker && docker-compose down

docker-logs:
	cd docker && docker-compose logs -f

# Development setup
dev-setup:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Profile CPU
profile-cpu:
	go tool pprof http://localhost:8080/debug/pprof/profile

# Profile memory
profile-mem:
	go tool pprof http://localhost:8080/debug/pprof/heap

# Watch for changes and rebuild
watch:
	reflex -s -r '\.go$$' -- make build

# Load test
load-test:
	go run cmd/loadtest/main.go

# Full CI pipeline
ci: fmt vet test lint sec

# Help target
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  run            - Build and run the server"
	@echo "  proto          - Generate protobuf files"
	@echo "  test           - Run tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  bench          - Run benchmarks"
	@echo "  test-integration - Run integration tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  lint           - Run linter"
	@echo "  sec            - Run security scan"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start services with Docker Compose"
	@echo "  docker-down    - Stop services"
	@echo "  docker-logs    - Show Docker logs"
	@echo "  dev-setup      - Setup development tools"
	@echo "  coverage       - Generate coverage report"
	@echo "  profile-cpu    - Profile CPU usage"
	@echo "  profile-mem    - Profile memory usage"
	@echo "  watch          - Watch for changes and rebuild"
	@echo "  load-test      - Run load test"
	@echo "  ci             - Run full CI pipeline"
	@echo "  help           - Show this help"
