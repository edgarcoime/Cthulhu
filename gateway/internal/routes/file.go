package routes

import (
	"github.com/edgarcoime/Cthulhu-gateway/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func FileRouter(app fiber.Router) {
	app.Post("/files/upload", handlers.UploadFile())
	app.Get("/files/s/:id", handlers.FileAccess())
	app.Get("/files/s/:id/d/:filename", handlers.FileDownload())
}
