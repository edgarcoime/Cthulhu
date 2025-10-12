module cthulhu-filemanager

go 1.25.1

require (
	cthulhu-shared v0.0.0
	github.com/joho/godotenv v1.5.1
)

require github.com/rabbitmq/amqp091-go v1.10.0 // indirect

replace cthulhu-shared => ../pkg
