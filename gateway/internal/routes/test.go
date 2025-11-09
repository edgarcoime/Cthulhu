package routes

import (
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

func TestRouter(app fiber.Router, s *services.Container) {
	app.Get("/test/services", ServicesFanoutTest(s))
}

func ServicesFanoutTest(s *services.Container) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}
