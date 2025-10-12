.PHONY: dev build run test lint clean stop

dev:
	go mod tidy
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && air

build:
	@mkdir -p tmp
	@mkdir -p bin
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

lint:
	golangci-lint run

stop:
	@echo "Stopping gateway development server..."
	@pkill -f "air" || true
	@pkill -f "bin/api" || true

clean:
	rm -rf tmp/ bin/
