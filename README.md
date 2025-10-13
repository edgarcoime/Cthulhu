# Cthulhu

## What is CTHULHU

## Table of contents <a name="toc"></a>

- [What is CTHULHU](#what-is-cthulhu)
- [Table of contents](#toc)
- [Prerequisites](#prerequisites)
- [Development Environment](#development-environment)
  - [1. Setup RabbitMQ](#1-setup-rabbitmq)
  - [2. Start by using root Makefile](#2-start-by-using-root-makefile)
  - [3. Test application](#3-test-application)
- [Docker](#docker)
  - [RabbitMQ](#rabbitmq)
  - [Client](#client)
  - [Gateway](#gateway)
  - [Filemanager](#filemanager)

## Prerequisites

- Golang 1.25.1 - [download](https://go.dev/doc/install)
- Node 22.22.0 - [download](https://nodejs.org/en/download)
- Docker 28.4.0 (Any recent version is fine) - [download](https://docs.docker.com/engine/install/)
- Make (Most UNIX based systems have it otherwise install using package manager)

## Development Environment

Start by navigating to the root of the project.

### 1. Setup RabbitMQ

```bash
# Navigate to rabbitmq folder
cd ./rabbitmq

# Build rabbitmq
docker build -t cthulhu-rabbitmq .

# Run the container with environment vars
docker run -d --name cthulhu-rabbitmq \
  -p 5672:5672 -p 15672:15672 -p 25672:25672 \
  cthulhu-rabbitmq

# Check if container is running
docker ps
```

### 2. Start by using root Makefile

The project is setup with a combination of make files
Navigate back to the root of the project with the root Makefile

```bash
# Look at root Make to see what is being run
cat Makefile

# Start dev environment
make dev
```

### 3. Test application

Test the main resources of the application.

1. Client (Navigate to [localhost:3000](http://localhost:3000))
2. Gateway (Navigate to [localhost:4000](http://localhost:4000))
3. Filemanager (Not fully implemented yet)

## Docker

All services can also be ran using Docker. This allows me the flexibility to later run orchestration through docker-compose or Kubernetes. Here are some ways the services can be run through docker.

### RabbitMQ

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

### Client

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

### Gateway

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

### Filemanager
Work in progress not fully implemented and working with Docker yet
