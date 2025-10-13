// Project naming conventions
// https://github.com/golang-standards/project-layout
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cthulhu-gateway/internal/pkg"
	"cthulhu-gateway/internal/routes"
	"cthulhu-gateway/pkg/file"
	"cthulhu-shared/rabbitmq"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Ensure fileDump directory exists
	if err := os.MkdirAll(pkg.FILE_FOLDER, 0o755); err != nil {
		log.Fatalf("Failed to create fileDump directory: %v", err)
	}

	// Initialize RabbitMQ manager
	rabbitMQManager := rabbitmq.NewManager(pkg.GetRabbitMQConfig())

	// Connect to RabbitMQ
	if err := rabbitMQManager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitMQManager.Close()

	// Start heartbeat monitoring
	rabbitMQManager.StartHeartbeat(ctx)

	// Declare response queue for receiving responses from filemanager
	responseQueue, err := rabbitMQManager.DeclareQueue("gateway.responses", true, false, false, false)
	if err != nil {
		log.Printf("Failed to declare response queue: %v", err)
	} else {
		log.Printf("Declared response queue: %s", responseQueue.Name)
	}

	// Start consuming responses from filemanager
	go startResponseConsumer(ctx, rabbitMQManager, responseQueue.Name)

	app := fiber.New()

	// Add Cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: pkg.CORS_ORIGIN,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Middleware
	app.Use(logger.New(logger.Config{
		// 2005-03-19 15:10:26,618 - simple_example - DEBUG - debug mess
		Format:     "${date} ${time},${pid} - ${ip}:${port} - ${status} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))
	app.Use(recover.New())

	// Initialize service with RabbitMQ dependency
	fileService := file.NewService(rabbitMQManager)

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	// File routes
	routes.FileRouter(app, fileService)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		app.Shutdown()
	}()

	if err := app.Listen(":" + pkg.PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// startResponseConsumer starts consuming responses from the filemanager
func startResponseConsumer(ctx context.Context, rabbitMQManager *rabbitmq.Manager, queueName string) {
	channel := rabbitMQManager.GetChannel()
	if channel == nil {
		log.Println("No active channel available for response consumption")
		return
	}

	// Set QoS to process one message at a time
	err := channel.Qos(1, 0, false)
	if err != nil {
		log.Printf("Failed to set QoS for response consumer: %v", err)
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
		log.Printf("Failed to register response consumer: %v", err)
		return
	}

	log.Printf("Started consuming responses from queue: %s", queueName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Response consumer context cancelled, stopping...")
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Response channel closed, stopping consumer...")
				return
			}

			// Process the response
			go processResponse(msg)
		}
	}
}

// processResponse processes a response from the filemanager
func processResponse(msg amqp.Delivery) {
	defer msg.Ack(false) // Acknowledge the message

	log.Printf("Received response from filemanager: %s", string(msg.Body))

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(msg.Body, &response); err != nil {
		log.Printf("Failed to parse response: %v", err)
		return
	}

	// Log the response details
	if status, ok := response["status"].(string); ok {
		log.Printf("Filemanager response status: %s", status)
	}
	if message, ok := response["message"].(string); ok {
		log.Printf("Filemanager response message: %s", message)
	}
	if processedAt, ok := response["processed_at"].(float64); ok {
		log.Printf("Response processed at: %d", int64(processedAt))
	}
}
