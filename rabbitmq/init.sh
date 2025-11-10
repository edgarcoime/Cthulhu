#!/bin/bash

# Initialize RabbitMQ with custom configuration
echo "Starting RabbitMQ initialization..."

# Start RabbitMQ in the background
rabbitmq-server &
RABBITMQ_PID=$!

# Wait for RabbitMQ to start
echo "Waiting for RabbitMQ to start..."
sleep 15

# Wait for RabbitMQ to be ready
until rabbitmqctl status > /dev/null 2>&1; do
    echo "Waiting for RabbitMQ to be ready..."
    sleep 2
done

# Enable management plugin
echo "Enabling management plugin..."
rabbitmq-plugins enable rabbitmq_management

# Enable tracing plugin (allows reading/tracing messages)
echo "Enabling tracing plugin..."
rabbitmq-plugins enable rabbitmq_tracing

# Restart RabbitMQ to apply plugin changes
echo "Restarting RabbitMQ to apply plugin changes..."
rabbitmqctl stop_app
rabbitmqctl start_app

# Wait for RabbitMQ to be ready again
sleep 5
until rabbitmqctl status > /dev/null 2>&1; do
    echo "Waiting for RabbitMQ to restart..."
    sleep 2
done

# Load definitions if they exist
if [ -f /etc/rabbitmq/definitions.json ]; then
    echo "Loading definitions..."
    rabbitmqctl import_definitions /etc/rabbitmq/definitions.json
fi

# Keep the container running
echo "RabbitMQ is ready with management and tracing plugins enabled!"
echo "Management UI: http://localhost:15672"
echo "Tracing UI: http://localhost:15672/#/traces"
wait $RABBITMQ_PID
