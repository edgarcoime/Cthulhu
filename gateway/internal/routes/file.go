package routes

import (
	"cthulhu-gateway/internal/handlers"
	"cthulhu-gateway/pkg/file"

	"github.com/gofiber/fiber/v2"
)

func FileRouter(app fiber.Router, service file.Service) {
	app.Post("/files/upload", handlers.UploadFile(service))
	app.Get("/files/s/:id", handlers.FileAccess(service))
	app.Get("/files/s/:id/d/:filename", handlers.FileDownload(service))
	app.Get("/test/rabbitmq", handlers.TestRabbitMQ(service))
}

/*
POST   /files/upload
GET    /files/s/:id
GET    /files/s/:id/d/:filename
GET    /files/s/:id/m - Access metadata, folder browse
DELETE /files/s/:id - Remove upload session

*/
