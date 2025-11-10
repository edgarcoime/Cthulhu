.PHONY: dev stop run build start test clean

dev:
	go mod tidy
	@if ! command -v air >/dev/null 2>&1; then \
		echo "Installing air for hot reloading..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && air

stop:
	@echo "Stopping filemanager development server..."
	@pkill -f "air" || true
	@pkill -f "bin/filemanager" || true

run:
	go run ./cmd/filemanager

build:
	@mkdir -p bin/ tmp/
	go build -o bin/filemanager ./cmd/filemanager

start:
	./bin/filemanager

test:
	go test -v ./...

clean:
	rm -rf tmp/ bin/

