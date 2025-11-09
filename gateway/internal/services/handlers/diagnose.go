package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/google/uuid"
)

// Handlers handle not just one service but its about the type.
// So in this case Diagnose wants to send messages to all services to see if they are all
// working properly
type DiagnoseHandler struct {
	manager *manager.Manager
	ctx     context.Context
}

func NewDiagnoseHandler(rmqManager *manager.Manager, ctx context.Context) *DiagnoseHandler {
	return &DiagnoseHandler{
		manager: rmqManager,
		ctx:     ctx,
	}
}

func (h *DiagnoseHandler) SetupQueuesAndBindings() error {
	// Declare topic exchange for diagnostic messages
	// Topic exchange allows flexible routing with wildcards
	err := h.manager.DeclareExchange(
		messages.DiagnoseExchange,
		"topic", // topic exchange type for flexible routing
		true,    // durable
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to declare diagnose exchange: %w", err)
	}

	return nil
}

// TestFanoutToAllServices sends a diagnostic message to all services to check if they're up
// Returns the transaction ID and any error
func (h *DiagnoseHandler) TestFanoutToAllServices() (string, error) {
	// Generate transaction ID
	transactionID := uuid.New().String()

	// Create diagnostic message
	diagnoseMsg := messages.DiagnoseMessage{
		TransactionID: transactionID,
		Operation:     "all", // Operation type: "all" means check if service is up
		Message:       "Health check - are you up?",
	}

	// Marshal to JSON
	messageBody, err := json.Marshal(diagnoseMsg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Ensure exchange is declared (idempotent)
	if err := h.manager.DeclareExchange(
		messages.DiagnoseExchange,
		"topic",
		true,
		false,
		false,
		false,
	); err != nil {
		return "", fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Publish message to the exchange
	// Services should bind with: diagnose.services.* to receive all diagnostic messages
	if err := h.manager.PublishMessage(
		h.ctx,
		messages.DiagnoseExchange,
		messages.TopicDiagnoseServicesAll, // routing key: diagnose.services.all
		"application/json",
		messageBody,
	); err != nil {
		return "", fmt.Errorf("failed to publish message: %w", err)
	}

	return transactionID, nil
}
