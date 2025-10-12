.PHONY: dev client gateway filemanager stop clean setup-env help

# Default development target - runs all services locally
dev: setup-env
	@echo "Starting complete development environment..."
	@echo "Starting Gateway, Filemanager, and Client..."
	@echo "Note: Make sure RabbitMQ is running first!"
	@echo "Press Ctrl+C to stop all services"
	@echo "----------------------------------------"
	@mkdir -p tmp
	@echo "Starting gateway..." > tmp/gateway.log
	@echo "Starting filemanager..." > tmp/filemanager.log
	@echo "Starting client..." > tmp/client.log
	@$(MAKE) -C gateway dev >> tmp/gateway.log 2>&1 &
	@sleep 2
	@$(MAKE) -C filemanager dev >> tmp/filemanager.log 2>&1 &
	@sleep 2
	@$(MAKE) -C client dev >> tmp/client.log 2>&1 &
	@sleep 2
	@echo "All services started. Showing combined logs:"
	@echo "----------------------------------------"
	@tail -f tmp/gateway.log tmp/filemanager.log tmp/client.log

# Start client development server
client.dev:
	$(MAKE) -C client dev

# Start gateway development server  
gateway.dev:
	$(MAKE) -C gateway dev

# Start filemanager development server
filemanager.dev:
	$(MAKE) -C filemanager dev

# Show combined logs
logs:
	@echo "Combined logs:"
	@tail -f tmp/gateway.log tmp/filemanager.log tmp/client.log

# Stop all services
stop:
	$(MAKE) -C client stop 2>/dev/null || true
	$(MAKE) -C gateway stop 2>/dev/null || true
	$(MAKE) -C filemanager stop 2>/dev/null || true
	@pkill -f "air" 2>/dev/null || true
	@pkill -f "next dev" 2>/dev/null || true
	@rm -rf tmp/

# Clean up everything
clean: stop
	$(MAKE) -C client clean
	$(MAKE) -C gateway clean
	$(MAKE) -C filemanager clean

# Setup environment files
setup-env:
	@echo "Setting up environment files..."
	@if [ ! -f client/.env.local ]; then \
		cp client/env.example client/.env.local; \
		echo "Created client/.env.local from env.example"; \
	fi
	@if [ ! -f gateway/.env ]; then \
		cp gateway/env.example gateway/.env; \
		echo "Created gateway/.env from env.example"; \
	fi
	@if [ ! -f filemanager/.env ]; then \
		cp filemanager/env.example filemanager/.env; \
		echo "Created filemanager/.env from env.example"; \
	fi

# Show help
help:
	@echo "Available commands:"
	@echo "  dev           - Start development environment (Gateway + Filemanager + Client)"
	@echo "  client.dev    - Start client development server only"
	@echo "  gateway.dev   - Start gateway development server only"
	@echo "  filemanager.dev - Start filemanager development server only"
	@echo "  logs          - Show combined logs from all services"
	@echo "  stop          - Stop all services"
	@echo "  clean         - Clean up build artifacts"
	@echo "  setup-env     - Copy environment files from examples"
	@echo "  help          - Show this help message"
	@echo ""
	@echo "Service URLs:"
	@echo "  Client:        http://localhost:3000"
	@echo "  Gateway API:   http://localhost:4000"
	@echo "  Filemanager:   http://localhost:5000"
	@echo ""
	@echo "Note: Start RabbitMQ manually before running 'make dev':"
	@echo "  cd rabbitmq && docker build -t cthulhu-rabbitmq . && docker run -d --name cthulhu-rabbitmq -p 5672:5672 -p 15672:15672 cthulhu-rabbitmq"
