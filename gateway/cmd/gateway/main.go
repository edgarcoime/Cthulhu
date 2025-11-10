package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarcoime/Cthulhu-gateway/internal/pkg"
	"github.com/edgarcoime/Cthulhu-gateway/internal/routes"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Start up requirements/configs
	ctx := context.Background()
	corsSettings := cors.New(cors.Config{
		AllowOrigins: pkg.CORS_ORIGIN,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	})
	customLogger := logger.New(logger.Config{
		// 2005-03-19 15:10:26,618 - simple_example - DEBUG - debug mess
		Format:     "${date} ${time},${pid} - ${ip}:${port} - ${status} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	})

	// ===== START SERVICES =====
	serviceContainer := services.NewContainer(ctx)
	defer func() {
		if err := serviceContainer.Shutdown(); err != nil {
			log.Printf("Error shutting down services: %v", err)
		}
	}()
	serviceContainer.StartListeners()

	// ===== START FIBER =====
	// Configure Fiber with increased body size limit for file uploads
	// Set to 100MB to allow for large file uploads
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024, // 500MB
	})

	// Add middleware
	app.Use(corsSettings)
	app.Use(customLogger)
	app.Use(recover.New())

	// ROUTES
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the gateway")
	})
	routes.FileRouter(app, serviceContainer)
	routes.TestRouter(app, serviceContainer)
	// SHUTDOWN GRACEFULLY
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down...")
		app.Shutdown()
	}()

	// Start server
	if err := app.Listen(":" + pkg.PORT); err != nil {
		log.Fatalf("Failed to strart server: %v\n", err)
	}
}
