PHONY: dev stop run build start test clean

dev:
	go mod tidy
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && air

stop:
	@echo "Stopping gateway development server..."
	@pkill -f "air" || true
	@pkill -f "bin/gateway" || true

run:
	go run ./cmd/gateway

build:
	@mkdir -p bin/ tmp/
	go build -o bin/gateway ./cmd/gateway

start:
	./bin/gateway

test:
	go test -v ./...

clean:
	rm -rf tmp/ bin/
