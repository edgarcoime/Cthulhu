.PHONY: dev build run test lint clean stop

dev:
	go mod tidy
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && air

build:
	go build -o filemanager ./cmd

run:
	go run ./cmd

test:
	go test ./...

lint:
	golangci-lint run

stop:
	@echo "Stopping filemanager development server..."
	@pkill -f "air" || true
	@pkill -f "filemanager" || true

clean:
	rm -f filemanager
