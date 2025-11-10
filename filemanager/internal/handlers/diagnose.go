package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ServiceName = "filemanager"
)

// HandleDiagnoseMessages processes diagnose messages and sends responses
func (h *Handler) HandleDiagnoseMessages(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		var diagnoseMsg messages.DiagnoseMessage
		if err := json.Unmarshal(msg.Body, &diagnoseMsg); err != nil {
			log.Printf("Failed to unmarshal diagnose message: %v", err)
			msg.Nack(false, false) // Don't requeue malformed messages
			continue
		}

		log.Printf("Received diagnose message: operation=%s, transaction_id=%s", diagnoseMsg.Operation, diagnoseMsg.TransactionID)

		// Create response
		response := messages.DiagnoseResponse{
			TransactionID: diagnoseMsg.TransactionID,
			ServiceName:   ServiceName,
			Operation:     diagnoseMsg.Operation,
			Status:        messages.DiagnoseStatusProcessed,
			Message:       "Filemanager service is operational",
			Data: map[string]interface{}{
				"service": ServiceName,
				"status":  "healthy",
			},
		}

		// Send response
		if err := h.sendDiagnoseResponse(response, &msg); err != nil {
			log.Printf("Failed to send diagnose response: %v", err)
			continue
		}

		// Acknowledge the message
		msg.Ack(false)
	}
}

// sendDiagnoseResponse publishes the diagnose response message
func (h *Handler) sendDiagnoseResponse(response messages.DiagnoseResponse, msg *amqp.Delivery) error {
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal diagnose response: %v", err)
		msg.Nack(false, true) // Requeue to retry
		return err
	}

	responseRoutingKey := fmt.Sprintf("%s.%s", messages.TopicDiagnoseServicesResponse, ServiceName)
	if err := h.manager.PublishMessage(
		h.ctx,
		messages.DiagnoseExchange,
		responseRoutingKey,
		"application/json",
		responseBody,
	); err != nil {
		log.Printf("Failed to publish diagnose response: %v", err)
		msg.Nack(false, true) // Requeue to retry
		return err
	}

	return nil
}
