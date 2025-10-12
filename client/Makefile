.PHONY: dev build start lint test clean stop

dev:
	npm install
	npm run dev

build:
	npm run build

start:
	npm run start

lint:
	npm run lint

test:
	npm test

stop:
	@echo "Stopping client development server..."
	@pkill -f "next dev" || true

clean:
	rm -rf .next node_modules
