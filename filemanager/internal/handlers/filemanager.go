package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	amqp "github.com/rabbitmq/amqp091-go"
)

// HandleFileManagerMessages processes filemanager operation messages
func (h *Handler) HandleFileManagerMessages(queueName string, msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		var request messages.FileManagerRequest
		if err := json.Unmarshal(msg.Body, &request); err != nil {
			log.Printf("Failed to unmarshal filemanager request: %v", err)
			msg.Nack(false, false) // Don't requeue malformed messages
			continue
		}

		log.Printf("Received filemanager request: queue=%s, transaction_id=%s", queueName, request.TransactionID)

		// Route to appropriate handler based on queue name
		var response messages.FileManagerResponse
		var err error

		switch queueName {
		case "filemanager.post.file":
			response, err = h.handlePostFile(request)
		case "filemanager.post.files":
			response, err = h.handlePostFiles(request)
		case "filemanager.get.file":
			response, err = h.handleGetFile(request)
		case "filemanager.get.files":
			response, err = h.handleGetFiles(request)
		case "filemanager.delete.file":
			response, err = h.handleDeleteFile(request)
		case "filemanager.delete.folder":
			response, err = h.handleDeleteFolder(request)
		default:
			err = fmt.Errorf("unknown queue: %s", queueName)
		}

		if err != nil {
			response = messages.FileManagerResponse{
				TransactionID: request.TransactionID,
				Success:       false,
				Error:         err.Error(),
			}
			log.Printf("Error handling request: %v", err)
		}

		// Send response
		if err := h.sendResponse(queueName, request.TransactionID, response, &msg); err != nil {
			log.Printf("Failed to send response: %v", err)
			continue
		}

		// Acknowledge the message
		msg.Ack(false)
	}
}

// sendResponse publishes the response message
func (h *Handler) sendResponse(queueName, transactionID string, response messages.FileManagerResponse, msg *amqp.Delivery) error {
	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %v", err)
		msg.Nack(false, true) // Requeue to retry
		return err
	}

	// Determine response routing key based on operation
	operation := getOperationFromQueue(queueName)
	responseRoutingKey := fmt.Sprintf("%s.%s.%s", messages.TopicFileManagerResponse, operation, transactionID)

	if err := h.manager.PublishMessage(
		h.ctx,
		messages.FileManagerExchange,
		responseRoutingKey,
		"application/json",
		responseBody,
	); err != nil {
		log.Printf("Failed to publish response: %v", err)
		msg.Nack(false, true) // Requeue to retry
		return err
	}

	return nil
}

// getOperationFromQueue extracts the operation name from queue name
func getOperationFromQueue(queueName string) string {
	// queueName format: "filemanager.post.file" -> "post.file"
	if len(queueName) > len("filemanager.") {
		return queueName[len("filemanager."):]
	}
	return queueName
}

// Handler methods for each operation
func (h *Handler) handlePostFile(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	// TODO: Implement file upload from message
	// For now, return an error indicating this needs file content
	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       false,
		Error:         "PostFile requires file content, which should be sent separately",
	}, nil
}

func (h *Handler) handlePostFiles(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	// TODO: Implement multiple file upload from message
	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       false,
		Error:         "PostFiles requires file content, which should be sent separately",
	}, nil
}

func (h *Handler) handleGetFile(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	if request.StorageID == "" || request.Filename == "" {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         "storage_id and filename are required",
		}, nil
	}

	// Get file from service
	fileReader, err := h.service.GetFile(h.ctx, request.TransactionID, request.StorageID, request.Filename)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}
	defer fileReader.Close()

	// TODO: For file content, we might need to send it via a different mechanism
	// For now, just return success
	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       true,
		StorageID:     request.StorageID,
		Data: map[string]interface{}{
			"filename": request.Filename,
			"note":     "File content should be retrieved via separate mechanism",
		},
	}, nil
}

func (h *Handler) handleGetFiles(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	if request.StorageID == "" {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         "storage_id is required",
		}, nil
	}

	// Get files from service
	fileInfos, err := h.service.GetFiles(h.ctx, request.TransactionID, request.StorageID)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}

	// Convert repository.FileInfo to messages.FileInfo
	files := make([]messages.FileInfo, len(fileInfos))
	var totalSize int64
	for i, fi := range fileInfos {
		files[i] = messages.FileInfo{
			Filename: fi.Filename,
			Size:     fi.Size,
		}
		totalSize += fi.Size
	}

	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       true,
		StorageID:     request.StorageID,
		Files:         files,
		TotalSize:     totalSize,
	}, nil
}

func (h *Handler) handleDeleteFile(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	if request.StorageID == "" || request.Filename == "" {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         "storage_id and filename are required",
		}, nil
	}

	err := h.service.DeleteFile(h.ctx, request.TransactionID, request.StorageID, request.Filename)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}

	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       true,
		StorageID:     request.StorageID,
	}, nil
}

func (h *Handler) handleDeleteFolder(request messages.FileManagerRequest) (messages.FileManagerResponse, error) {
	if request.StorageID == "" {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         "storage_id is required",
		}, nil
	}

	err := h.service.DeleteFolder(h.ctx, request.TransactionID, request.StorageID)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}

	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       true,
		StorageID:     request.StorageID,
	}, nil
}
