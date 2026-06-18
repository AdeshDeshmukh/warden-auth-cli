.PHONY: build run test test-cover docker-up docker-down lint clean help

build:
	go build -o bin/warden ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test -v -race ./tests/...

test-cover:
	go test -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down -v

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out coverage.html

help:
	@echo ""
	@echo "  Warden Auth CLI — Available Make Targets"
	@echo ""
	@echo "  build        Compile binary to bin/warden"
	@echo "  run          Run locally with go run"
	@echo "  test         Run all tests with race detector"
	@echo "  test-cover   Generate HTML coverage report"
	@echo "  docker-up    Build and start with docker-compose"
	@echo "  docker-down  Stop containers and remove volumes"
	@echo "  lint         Run go vet"
	@echo "  clean        Remove build artifacts"
	@echo ""