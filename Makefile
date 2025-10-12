.PHONY: dev client gateway stop clean setup-env help

# Default development target - runs client and gateway locally
dev: setup-env
	@echo "Starting development environment..."
	@echo "Starting both client and gateway with combined logs..."
	@echo "Press Ctrl+C to stop both services"
	@echo "----------------------------------------"
	@mkdir -p tmp
	@echo "Starting gateway..." > tmp/gateway.log
	@echo "Starting client..." > tmp/client.log
	@$(MAKE) -C gateway dev >> tmp/gateway.log 2>&1 &
	@sleep 3
	@$(MAKE) -C client dev >> tmp/client.log 2>&1 &
	@sleep 2
	@echo "Both services started. Showing combined logs:"
	@echo "----------------------------------------"
	@tail -f tmp/gateway.log tmp/client.log

# Start client development server
client.dev:
	$(MAKE) -C client dev

# Start gateway development server  
gateway.dev:
	$(MAKE) -C gateway dev

# Show combined logs
logs:
	@echo "Combined logs:"
	@tail -f tmp/gateway.log tmp/client.log

# Stop all services
stop:
	$(MAKE) -C client stop 2>/dev/null || true
	$(MAKE) -C gateway stop 2>/dev/null || true
	@pkill -f "air" 2>/dev/null || true
	@pkill -f "next dev" 2>/dev/null || true
	@rm -rf tmp/

# Clean up everything
clean: stop
	$(MAKE) -C client clean
	$(MAKE) -C gateway clean

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

# Show help
help:
	@echo "Available commands:"
	@echo "  dev        - Start development environment with combined logs"
	@echo "  client.dev - Start client development server"
	@echo "  gateway.dev- Start gateway development server"
	@echo "  logs       - Show combined logs from both services"
	@echo "  stop       - Stop all services"
	@echo "  clean      - Clean up build artifacts"
	@echo "  setup-env  - Copy environment files from examples"
	@echo "  help       - Show this help message"
