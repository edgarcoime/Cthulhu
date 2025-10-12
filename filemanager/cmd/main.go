package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cthulhu-filemanager/internal/pkg"
	"cthulhu-shared/rabbitmq"
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

	// Example: Declare a queue
	queue, err := rabbitMQManager.DeclareQueue("filemanager.queue", true, false, false, false)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
	} else {
		log.Printf("Declared queue: %s", queue.Name)
	}

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
