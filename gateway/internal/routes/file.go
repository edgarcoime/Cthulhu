package routes

import (
	"cthulhu-gateway/internal/handlers"
	"cthulhu-gateway/pkg/file"

	"github.com/gofiber/fiber/v2"
)

func FileRouter(app fiber.Router, service file.Service) {
	app.Post("/upload", handlers.UploadFile(service))
	app.Get("/files/:id", handlers.FileAccess(service))
	app.Get("/files/:id/download/:filename", handlers.FileDownload(service))
}
