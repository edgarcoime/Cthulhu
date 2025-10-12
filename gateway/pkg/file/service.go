package file

import "cthulhu-gateway/internal/pkg"

type Service interface {
	// Add methods that require RabbitMQ functionality
	PublishFileEvent(eventType string, data []byte) error
}

type service struct {
	rabbitMQ *pkg.RabbitMQManager
}

// NewService creates a new file service with RabbitMQ dependency
func NewService(rabbitMQ *pkg.RabbitMQManager) Service {
	return &service{
		rabbitMQ: rabbitMQ,
	}
}

// PublishFileEvent publishes a file-related event to RabbitMQ
func (s *service) PublishFileEvent(eventType string, data []byte) error {
	// Example implementation - you can customize this based on your needs
	return s.rabbitMQ.PublishMessage(nil, "", "file.events", data)
}
