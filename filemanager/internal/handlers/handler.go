package handlers

import (
	"context"

	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
)

// Handler holds dependencies for message handlers
type Handler struct {
	service service.Service
	manager *manager.Manager
	ctx     context.Context
}

// NewHandler creates a new handler instance
func NewHandler(service service.Service, manager *manager.Manager, ctx context.Context) *Handler {
	return &Handler{
		service: service,
		manager: manager,
		ctx:     ctx,
	}
}
