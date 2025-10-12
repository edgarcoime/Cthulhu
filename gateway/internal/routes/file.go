package routes

import (
	"cthulhu-gateway/pkg/file"

	"github.com/gofiber/fiber/v2"
)

func FileRouter(app fiber.Router, service file.Service) {
	app.Get("/file/upload")
}
