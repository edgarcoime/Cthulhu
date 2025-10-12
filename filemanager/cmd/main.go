package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cthulhu-filemanager/internal/pkg"
	"cthulhu-shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize RabbitMQ manager using shared package
	rabbitMQManager := rabbitmq.NewManager(pkg.GetRabbitMQConfig())

	// Connect to RabbitMQ with retry logic
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := rabbitMQManager.Connect(ctx); err != nil {
			log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v", i+1, maxRetries, err)
			if i == maxRetries-1 {
				log.Fatal("Max connection retries reached, exiting")
			}
			time.Sleep(time.Duration(i+1) * 2 * time.Second) // Exponential backoff
			continue
		}
		break
	}
	defer rabbitMQManager.Close()

	// Start heartbeat monitoring
	rabbitMQManager.StartHeartbeat(ctx)

	log.Println("Filemanager service connected to RabbitMQ successfully!")

	// Declare test queue for interop testing
	testQueue, err := rabbitMQManager.DeclareQueue("filemanager.test", true, false, false, false)
	if err != nil {
		log.Printf("Failed to declare test queue: %v", err)
	} else {
		log.Printf("Declared test queue: %s", testQueue.Name)
	}

	// Start consuming messages from the test queue
	go startTestMessageConsumer(ctx, rabbitMQManager, testQueue.Name)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Keep the service running until interrupted
	select {
	case <-sigChan:
		log.Println("Received shutdown signal, shutting down gracefully...")
		cancel()
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	}

	// Give some time for cleanup
	time.Sleep(1 * time.Second)
	log.Println("Filemanager service stopped")
}

// startTestMessageConsumer starts consuming messages from the test queue
func startTestMessageConsumer(ctx context.Context, rabbitMQManager *rabbitmq.Manager, queueName string) {
	channel := rabbitMQManager.GetChannel()
	if channel == nil {
		log.Println("No active channel available for message consumption")
		return
	}

	// Set QoS to process one message at a time
	err := channel.Qos(1, 0, false)
	if err != nil {
		log.Printf("Failed to set QoS: %v", err)
		return
	}

	// Start consuming messages
	msgs, err := channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	log.Printf("Started consuming messages from queue: %s", queueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Message consumer context cancelled, stopping...")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed, stopping consumer...")
				return
			}

			// Process the message
			go processTestMessage(ctx, rabbitMQManager, msg)
		}
	}
}

// processTestMessage processes a test message and responds after 2 seconds
func processTestMessage(ctx context.Context, rabbitMQManager *rabbitmq.Manager, msg amqp.Delivery) {
	defer msg.Ack(false) // Acknowledge the message

	log.Printf("Received test message: %s", string(msg.Body))

	// Parse the incoming message
	var incomingMsg map[string]interface{}
	if err := json.Unmarshal(msg.Body, &incomingMsg); err != nil {
		log.Printf("Failed to parse incoming message: %v", err)
		return
	}

	// Wait for 2 seconds as requested
	log.Println("Processing message... waiting 2 seconds")
	time.Sleep(2 * time.Second)

	// Create response message
	response := map[string]interface{}{
		"status":       "processed",
		"message":      "Hello back from Filemanager!",
		"processed_at": time.Now().Unix(),
		"original":     incomingMsg,
		"response_id":  fmt.Sprintf("resp_%d", time.Now().UnixNano()),
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		return
	}

	// Send response back to gateway (using a different queue for responses)
	responseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = rabbitMQManager.PublishMessage(responseCtx, "", "gateway.responses", responseBytes)
	if err != nil {
		log.Printf("Failed to send response: %v", err)
	} else {
		log.Printf("Response sent successfully: %s", string(responseBytes))
	}
}
