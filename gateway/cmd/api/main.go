// Project naming conventions
// https://github.com/golang-standards/project-layout
package main

import (
	"context"
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
