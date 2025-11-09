package messages

// TestFanoutMessage represents a fanout message sent to all services for testing
// Topic: test.services
type TestFanoutMessage struct {
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

// TestFanoutResponseStatus represents the status of a test fanout response
type TestFanoutResponseStatus string

const (
	// TestFanoutStatusReceived indicates the message was received
	TestFanoutStatusReceived TestFanoutResponseStatus = "received"

	// TestFanoutStatusProcessed indicates the message was processed successfully
	TestFanoutStatusProcessed TestFanoutResponseStatus = "processed"

	// TestFanoutStatusError indicates an error occurred while processing
	TestFanoutStatusError TestFanoutResponseStatus = "error"
)

// TestFanoutResponse represents a response from a service after receiving a test fanout message
// Topic: test.services.response
type TestFanoutResponse struct {
	TransactionID string                   `json:"transaction_id"`
	ServiceName   string                   `json:"service_name"`
	Status        TestFanoutResponseStatus `json:"status"`
	Message       string                   `json:"message,omitempty"` // Optional response message
}
