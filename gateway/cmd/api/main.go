// Project naming conventions
// https://github.com/golang-standards/project-layout
package main

import (
	"log"
	"os"

	"cthulhu-gateway/internal/pkg"
	"cthulhu-gateway/internal/routes"
	"cthulhu-gateway/pkg/file"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Ensure fileDump directory exists
	if err := os.MkdirAll(pkg.FILE_FOLDER, 0o755); err != nil {
		log.Fatalf("Failed to create fileDump directory: %v", err)
	}

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

	// Initialize service
	var fileService file.Service

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	// File routes
	routes.FileRouter(app, fileService)

	if err := app.Listen(":" + pkg.PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
