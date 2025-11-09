package handlers

import (
	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

// ServicesFanoutTest handles the HTTP request for testing fanout to all services
func ServicesFanoutTest(s *services.Container) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Call the service handler to perform the business logic
		transactionID, err := s.DiagnoseHandler.TestFanoutToAllServices()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Return success response
		return c.JSON(fiber.Map{
			"success":        true,
			"transaction_id": transactionID,
			"message":        "Diagnostic message sent successfully to all services",
			"topic":          messages.TopicDiagnoseServicesAll,
		})
	}
}
