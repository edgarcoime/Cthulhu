package messages

// DiagnoseMessage represents a diagnostic message sent to services
// Topic: diagnose.services.<operation>
type DiagnoseMessage struct {
	TransactionID string `json:"transaction_id"`
	Operation     string `json:"operation"` // e.g., "health", "status", "load", "all"
	Message       string `json:"message,omitempty"`
}

// DiagnoseResponseStatus represents the status of a diagnostic response
type DiagnoseResponseStatus string

const (
	// DiagnoseStatusReceived indicates the message was received
	DiagnoseStatusReceived DiagnoseResponseStatus = "received"

	// DiagnoseStatusProcessed indicates the message was processed successfully
	DiagnoseStatusProcessed DiagnoseResponseStatus = "processed"

	// DiagnoseStatusError indicates an error occurred while processing
	DiagnoseStatusError DiagnoseResponseStatus = "error"
)

// DiagnoseResponse represents a response from a service after receiving a diagnostic message
// Topic: diagnose.services.response.<service-name>
type DiagnoseResponse struct {
	TransactionID string                 `json:"transaction_id"`
	ServiceName   string                 `json:"service_name"`
	Operation     string                 `json:"operation"` // The operation that was requested
	Status        DiagnoseResponseStatus `json:"status"`
	Message       string                 `json:"message,omitempty"` // Optional response message
	Data          map[string]interface{} `json:"data,omitempty"`    // Optional operation-specific data
}
