package pkg

import "cthulhu-shared/rabbitmq"

var (
	FILE_FOLDER = GetEnv("FILE_FOLDER", "./app/fileDump")
	PORT        = GetEnv("PORT", "4000")
	CORS_ORIGIN = GetEnv("CORS_ORIGIN", "http://localhost:3000")

	// AMQP config
	AMPQ_USER  = GetEnv("AMQP_USER", "guest")
	AMPQ_PASS  = GetEnv("AMQP_PASS", "guest")
	AMPQ_HOST  = GetEnv("AMQP_HOST", "localhost")
	AMPQ_PORT  = GetEnv("AMQP_PORT", "5672")
	AMPQ_VHOST = GetEnv("AMQP_VHOST", "/")
)

// GetRabbitMQConfig returns RabbitMQ configuration
func GetRabbitMQConfig() rabbitmq.Config {
	return rabbitmq.Config{
		User:     AMPQ_USER,
		Password: AMPQ_PASS,
		Host:     AMPQ_HOST,
		Port:     AMPQ_PORT,
		VHost:    AMPQ_VHOST,
	}
}
