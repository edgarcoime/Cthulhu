package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/handlers"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
)

const (
	ServiceName = "filemanager"
)

type RMQServerConfig struct {
	User           string
	Password       string
	Host           string
	Port           string
	VHost          string
	ConnectionName string
}

type rmqServer struct {
	handler *handlers.Handler
	manager *manager.Manager
}

// ListenRMQ starts the RabbitMQ server and listens for messages
func ListenRMQ(s service.Service, cfg *RMQServerConfig) {
	ctx := context.Background()

	// Create RabbitMQ Manager
	rmqManager := manager.NewManager(manager.Config{
		User:           cfg.User,
		Password:       cfg.Password,
		Host:           cfg.Host,
		Port:           cfg.Port,
		VHost:          cfg.VHost,
		ConnectionName: cfg.ConnectionName,
	})

	// Connect to RabbitMQ
	if err := rmqManager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	// Start heartbeat monitoring
	rmqManager.StartHeartbeat(ctx)

	// Create handler instance
	handler := handlers.NewHandler(s, rmqManager, ctx)

	// Create server instance
	server := &rmqServer{
		handler: handler,
		manager: rmqManager,
	}

	// Setup exchanges, queues, and bindings
	if err := server.setupExchangesAndQueues(); err != nil {
		log.Fatalf("Failed to setup exchanges and queues: %v", err)
	}

	// Start consuming messages
	if err := server.startConsumers(); err != nil {
		log.Fatalf("Failed to start consumers: %v", err)
	}

	log.Println("Filemanager RabbitMQ server started and listening for messages")

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down filemanager server...")
	if err := rmqManager.Close(); err != nil {
		log.Printf("Error closing RabbitMQ connection: %v", err)
	}
}

// setupExchangesAndQueues declares exchanges and sets up queues with bindings
func (s *rmqServer) setupExchangesAndQueues() error {
	// Declare diagnose exchange
	if err := s.manager.DeclareExchange(
		messages.DiagnoseExchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
	); err != nil {
		return fmt.Errorf("failed to declare diagnose exchange: %w", err)
	}

	// Declare filemanager exchange
	if err := s.manager.DeclareExchange(
		messages.FileManagerExchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
	); err != nil {
		return fmt.Errorf("failed to declare filemanager exchange: %w", err)
	}

	// Create queue for diagnose messages (bind to all diagnose operations)
	diagnoseQueue, err := s.manager.DeclareQueue(
		fmt.Sprintf("%s.diagnose", ServiceName),
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to declare diagnose queue: %w", err)
	}

	// Bind diagnose queue to exchange (listen to all diagnose operations)
	if err := s.manager.QueueBind(
		diagnoseQueue.Name,
		"diagnose.services.*", // Listen to all diagnose operations
		messages.DiagnoseExchange,
		false, // no-wait
	); err != nil {
		return fmt.Errorf("failed to bind diagnose queue: %w", err)
	}

	// Create queues for filemanager operations
	fileManagerQueues := []struct {
		name       string
		routingKey string
	}{
		{"filemanager.post.file", messages.TopicFileManagerPostFile},
		{"filemanager.post.file.chunk", messages.TopicFileManagerPostFileChunk},
		{"filemanager.post.files", messages.TopicFileManagerPostFiles},
		{"filemanager.get.file", messages.TopicFileManagerGetFile},
		{"filemanager.get.files", messages.TopicFileManagerGetFiles},
		{"filemanager.delete.file", messages.TopicFileManagerDeleteFile},
		{"filemanager.delete.folder", messages.TopicFileManagerDeleteFolder},
	}

	for _, q := range fileManagerQueues {
		queue, err := s.manager.DeclareQueue(
			q.name,
			true,  // durable
			false, // auto-delete
			false, // exclusive
			false, // no-wait
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.name, err)
		}

		if err := s.manager.QueueBind(
			queue.Name,
			q.routingKey,
			messages.FileManagerExchange,
			false, // no-wait
		); err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", q.name, err)
		}
	}

	return nil
}

// startConsumers starts consuming messages from all queues
func (s *rmqServer) startConsumers() error {
	// Start diagnose consumer
	diagnoseQueue := fmt.Sprintf("%s.diagnose", ServiceName)
	diagnoseMsgs, err := s.manager.Consume(
		diagnoseQueue,
		"",    // consumer tag (empty = auto-generate)
		false, // auto-ack (false = manual ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to start diagnose consumer: %w", err)
	}

	go s.handler.HandleDiagnoseMessages(diagnoseMsgs)

	// Start filemanager operation consumers
	fileManagerQueues := []string{
		"filemanager.post.file",
		"filemanager.post.file.chunk",
		"filemanager.post.files",
		"filemanager.get.file",
		"filemanager.get.files",
		"filemanager.delete.file",
		"filemanager.delete.folder",
	}

	for _, queueName := range fileManagerQueues {
		msgs, err := s.manager.Consume(
			queueName,
			"",    // consumer tag
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
		)
		if err != nil {
			return fmt.Errorf("failed to start consumer for %s: %w", queueName, err)
		}

		go s.handler.HandleFileManagerMessages(queueName, msgs)
	}

	return nil
}
