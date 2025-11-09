package services

import (
	"context"
	"log"

	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services/handlers"
)

// TODO: create interface to fetch handlers

type Container struct {
	RMQManager      *manager.Manager
	Ctx             context.Context
	DiagnoseHandler *handlers.DiagnoseHandler
}

func NewContainer(ctx context.Context) *Container {
	// Create RabbitMQ Manager
	rmqManager := manager.NewManager(manager.Config{
		// TODO: Change to env
		User:           "guest",
		Password:       "guest",
		Host:           "localhost",
		Port:           "5672",
		VHost:          "/",
		ConnectionName: "gateway",
	})

	// Connect to rmq
	if err := rmqManager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// Start heartbeat monitoring
	rmqManager.StartHeartbeat(ctx)

	// Create service handlers that will be communicating with
	diagnoseHandler := handlers.NewDiagnoseHandler(rmqManager, ctx)

	// setup queues and bindings
	if err := diagnoseHandler.SetupQueuesAndBindings(); err != nil {
		log.Fatalf("Failed to setup queues and bindings: %v", err)
	}

	return &Container{
		RMQManager:      rmqManager,
		Ctx:             ctx,
		DiagnoseHandler: diagnoseHandler,
	}
}

// Start listeners for each service
func (c *Container) StartListeners() {
	log.Println("All events listners started")
}

// Shutdown all connections
func (c *Container) Shutdown() error {
	log.Println("Shutting down services...")
	c.RMQManager.Close()
	return nil
}
