package pkg

import (
	"github.com/edgarcoime/Cthulhu-common/pkg/env"
)

var (
	// Storage Configuration
	STORAGE_PATH = env.GetEnv("STORAGE_PATH", "/tmp/fileDump")

	// RabbitMQ Configuration
	AMQP_USER  = env.GetEnv("AMQP_USER", "guest")
	AMQP_PASS  = env.GetEnv("AMQP_PASS", "guest")
	AMQP_HOST  = env.GetEnv("AMQP_HOST", "localhost")
	AMQP_PORT  = env.GetEnv("AMQP_PORT", "5672")
	AMQP_VHOST = env.GetEnv("AMQP_VHOST", "/")

	// Logging
	LOG_LEVEL = env.GetEnv("LOG_LEVEL", "info")
)
