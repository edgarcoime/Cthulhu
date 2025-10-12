.PHONY: dev build run test lint clean stop

dev:
	go mod tidy
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && air

build:
	go build -o gateway

run:
	go run main.go

lint:
	golangci-lint run

stop:
	@echo "Stopping gateway development server..."
	@pkill -f "air" || true
	@pkill -f "bin/api" || true

clean:
	rm -f gateway bin/api
