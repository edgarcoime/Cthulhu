package messages

// Topic constants for RabbitMQ routing
// Using hierarchical structure: diagnose.services.<operation>
const (
	// DiagnoseExchange is the exchange name for diagnostic messages
	DiagnoseExchange = "diagnose"

	// TopicDiagnoseServicesAll sends diagnostic messages to all services
	// Services should bind with: diagnose.services.*
	TopicDiagnoseServicesAll = "diagnose.services.all"

	// TopicDiagnoseServicesHealth is for health check requests
	TopicDiagnoseServicesHealth = "diagnose.services.health"

	// TopicDiagnoseServicesStatus is for service status requests
	TopicDiagnoseServicesStatus = "diagnose.services.status"

	// TopicDiagnoseServicesLoad is for service load/metrics requests
	TopicDiagnoseServicesLoad = "diagnose.services.load"

	// TopicDiagnoseServicesResponse is the topic for diagnostic responses (responses from services)
	// Format: diagnose.services.response.<service-name>
	TopicDiagnoseServicesResponse = "diagnose.services.response"
)
