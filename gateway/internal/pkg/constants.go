package pkg

var (
	FILE_FOLDER = GetEnv("FILE_FOLDER", "./app/fileDump")
	PORT        = GetEnv("PORT", "4000")
	CORS_ORIGIN = GetEnv("CORS_ORIGIN", "http://localhost:3000")

	// AMQP config
	AMPQ_USER = GetEnv("AMQP_USER", "guest")
	AMPQ_PASS = GetEnv("AMQP_PASS", "guest")
	AMPQ_HOST = GetEnv("AMQP_HOST", "localhost")
	AMPQ_PORT = GetEnv("AMQP_PORT", "5672")
)
