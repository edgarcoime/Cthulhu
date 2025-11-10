package pkg

import (
	"github.com/edgarcoime/Cthulhu-common/pkg/env"
)

var (
	FILE_FOLDER = env.GetEnv("FILE_FOLDER", "./app/fileDump")
	PORT        = env.GetEnv("PORT", "4000")
	CORS_ORIGIN = env.GetEnv("CORS_ORIGIN", "http://localhost:3000")

	// AMQP config
	AMPQ_USER  = env.GetEnv("AMQP_USER", "guest")
	AMPQ_PASS  = env.GetEnv("AMQP_PASS", "guest")
	AMPQ_HOST  = env.GetEnv("AMQP_HOST", "localhost")
	AMPQ_PORT  = env.GetEnv("AMQP_PORT", "5672")
	AMPQ_VHOST = env.GetEnv("AMQP_VHOST", "/")
)
