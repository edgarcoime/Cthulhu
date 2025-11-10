package routes

import (
	"github.com/edgarcoime/Cthulhu-gateway/internal/handlers"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

func FileRouter(app fiber.Router, services *services.Container) {
	// OLD version
	// app.Post("/files/upload", handlers.UploadFile())
	// app.Get("/files/s/:id", handlers.FileAccess())
	// app.Get("/files/s/:id/d/:filename", handlers.FileDownload())

	// new
	app.Post("/files/upload", handlers.RMQFileUpload(services))
	app.Get("/files/s/:id", handlers.RMQFileAccess(services))
	app.Get("/files/s/:id/d/:filename", handlers.RMQFileDownload(services))
}
