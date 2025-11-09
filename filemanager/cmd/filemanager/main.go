package main

import (
	"log"

	"github.com/edgarcoime/Cthulhu-filemanager/internal/pkg"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/repository"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/server"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
)

// TODO: Add air.toml config for hot reloading
func main() {
	// Initialize repository
	r, err := repository.NewLocalRepository(pkg.STORAGE_PATH)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer r.Close()

	// Initialize service
	s := service.NewFileManagerService(r)

	// Configure RabbitMQ server
	cfg := &server.RMQServerConfig{
		User:           pkg.AMQP_USER,
		Password:       pkg.AMQP_PASS,
		Host:           pkg.AMQP_HOST,
		Port:           pkg.AMQP_PORT,
		VHost:          pkg.AMQP_VHOST,
		ConnectionName: "filemanager",
	}

	// Start RabbitMQ server
	server.ListenRMQ(s, cfg)
}
