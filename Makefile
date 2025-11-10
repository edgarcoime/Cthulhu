.PHONY: dev stop clean

# Start all services
dev:
	@trap '$(MAKE) stop; exit' INT TERM; \
	echo "Starting services..."; \
	mkdir -p gateway/tmp/logs filemanager/tmp/logs client/tmp/logs tmp/pids; \
	(cd gateway && $(MAKE) dev > ../gateway/tmp/logs/gateway.log 2>&1) & \
		echo $$! > tmp/pids/gateway.pid; \
	sleep 2; \
	(cd filemanager && $(MAKE) dev > ../filemanager/tmp/logs/filemanager.log 2>&1) & \
		echo $$! > tmp/pids/filemanager.pid; \
	sleep 2; \
	(cd client && $(MAKE) dev > ../client/tmp/logs/client.log 2>&1) & \
		echo $$! > tmp/pids/client.pid; \
	sleep 2; \
	echo "All services started!"; \
	tail -f gateway/tmp/logs/gateway.log filemanager/tmp/logs/filemanager.log client/tmp/logs/client.log

# Stop all services
stop:
	@echo "Stopping all services..."
	@if [ -f tmp/pids/gateway.pid ]; then \
		pid=$$(cat tmp/pids/gateway.pid); \
		if ps -p $$pid > /dev/null 2>&1; then \
			pkill -P $$pid 2>/dev/null || true; \
			kill $$pid 2>/dev/null || true; \
		fi; \
		rm -f tmp/pids/gateway.pid; \
	fi
	@if [ -f tmp/pids/filemanager.pid ]; then \
		pid=$$(cat tmp/pids/filemanager.pid); \
		if ps -p $$pid > /dev/null 2>&1; then \
			pkill -P $$pid 2>/dev/null || true; \
			kill $$pid 2>/dev/null || true; \
		fi; \
		rm -f tmp/pids/filemanager.pid; \
	fi
	@if [ -f tmp/pids/client.pid ]; then \
		pid=$$(cat tmp/pids/client.pid); \
		if ps -p $$pid > /dev/null 2>&1; then \
			pkill -P $$pid 2>/dev/null || true; \
			kill $$pid 2>/dev/null || true; \
		fi; \
		rm -f tmp/pids/client.pid; \
	fi
	@pkill -f "air" 2>/dev/null || true
	@pkill -f "next dev" 2>/dev/null || true
	@echo "All services stopped!"

# Clean up temporary files
clean: stop
	@echo "Cleaning up..."
	@rm -rf gateway/tmp gateway/bin
	@rm -rf filemanager/tmp filemanager/bin
	@rm -rf client/.next
	@rm -rf tmp/
	@echo "Cleanup complete!"
