package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarcoime/Cthulhu-gateway/internal/pkg"
	"github.com/edgarcoime/Cthulhu-gateway/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
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

	// INITIALIZE SERVICES HERE

	// ROUTES
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the gateway")
	})
	routes.FileRouter(app)

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
