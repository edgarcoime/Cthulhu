package messages

// Topic constants for RabbitMQ routing
const (
	// TopicTestServices is the topic for test fanout messages (sent to all services)
	TopicTestServices = "test.services"
	
	// TopicTestServicesResponse is the topic for test fanout responses (responses from services)
	TopicTestServicesResponse = "test.services.response"
)

