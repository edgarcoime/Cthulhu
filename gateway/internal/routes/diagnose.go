package routes

import (
	"github.com/edgarcoime/Cthulhu-gateway/internal/handlers"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

func TestRouter(app fiber.Router, s *services.Container) {
	app.Get("/diagnose/services/all", handlers.ServicesFanoutTest(s))
}
