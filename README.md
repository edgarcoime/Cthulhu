# Cthulhu

## Client

### Docker: Build and run

Run the client by itself without Docker-compose orchestration

```bash
# Build Docker image
cd ./client
docker build -t cthulhu-client .

# Run with environment variable overrides
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_APP_NAME=Cthulhu \
  -e NEXT_PUBLIC_VERSION=0.1.0 \
  -e ROOT_DOMAIN=http://localhost:3000 \
  cthulhu-client
```

## Gateway

### Docker: Build and run

Run the gateway by itself without Docker-compose orchestration

```bash
# Build docker image
cd ./gateway
docker build -t cthulhu-gateway .

# Run with environment variable overrides
docker run -p 4000:4000 \
  -e PORT=4000 \
  -e FILE_FOLDER=/app/fileDump \
  -e CORS_ORIGIN=http://localhost:3000 \
  -e LOG_LEVEL=info \
  cthulhu-gateway

```

## Rabbitmq

### Docker: Build and run

Run the rabbitmq

```bash
cd ./rabbitmq
docker build -t cthulhu-rabbitmq

# Run
docker run -d \
  --name cthulhu-rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  -p 25672:25672 \
  cthulhu-rabbitmq
```
