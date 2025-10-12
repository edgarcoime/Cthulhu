#!/bin/bash

# Initialize RabbitMQ with custom configuration
echo "Starting RabbitMQ initialization..."

# Start RabbitMQ in the background
rabbitmq-server &

# Wait for RabbitMQ to start
echo "Waiting for RabbitMQ to start..."
sleep 10

# Enable management plugin
echo "Enabling management plugin..."
rabbitmq-plugins enable rabbitmq_management

# Load definitions if they exist
if [ -f /etc/rabbitmq/definitions.json ]; then
    echo "Loading definitions..."
    rabbitmqctl import_definitions /etc/rabbitmq/definitions.json
fi

# Keep the container running
echo "RabbitMQ is ready!"
wait
