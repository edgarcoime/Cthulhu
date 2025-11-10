package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ChunkSize is the maximum size of a chunk in bytes (1MB)
// This ensures messages stay well under RabbitMQ's practical limits
const ChunkSize = 1024 * 1024 // 1MB

// HandleFileManagerMessages processes filemanager operation messages
func (h *Handler) HandleFileManagerMessages(queueName string, msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		log.Printf("Received filemanager message: queue=%s", queueName)

		// Route to appropriate handler based on queue name
		var response messages.FileManagerResponse
		var err error

		switch queueName {
		case "filemanager.post.file.chunk":
			// Handle chunk message
			var chunkRequest messages.FileChunkRequest
			if err := json.Unmarshal(msg.Body, &chunkRequest); err != nil {
				log.Printf("Failed to unmarshal chunk request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handleFileChunk(chunkRequest, &msg)
		case "filemanager.post.file":
			// Try to unmarshal as FileUploadRequest (with content) first
			var uploadRequest messages.FileUploadRequest
			if unmarshalErr := json.Unmarshal(msg.Body, &uploadRequest); unmarshalErr == nil && uploadRequest.Content != "" {
				// Successfully unmarshaled as FileUploadRequest with content
				response, err = h.handlePostFileWithContent(uploadRequest)
			} else {
				// Fall back to FileManagerRequest (without content)
				var request messages.FileManagerRequest
				if unmarshalErr := json.Unmarshal(msg.Body, &request); unmarshalErr != nil {
					log.Printf("Failed to unmarshal filemanager request: %v", unmarshalErr)
					msg.Nack(false, false) // Don't requeue malformed messages
					continue
				}
				response, err = h.handlePostFile(request)
			}
		case "filemanager.post.files":
			var request messages.FileManagerRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("Failed to unmarshal filemanager request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handlePostFiles(request)
		case "filemanager.get.file":
			var request messages.FileManagerRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("Failed to unmarshal filemanager request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handleGetFile(request)
		case "filemanager.get.files":
			var request messages.FileManagerRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("Failed to unmarshal filemanager request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handleGetFiles(request)
		case "filemanager.delete.file":
			var request messages.FileManagerRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("Failed to unmarshal filemanager request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handleDeleteFile(request)
		case "filemanager.delete.folder":
			var request messages.FileManagerRequest
			if err := json.Unmarshal(msg.Body, &request); err != nil {
				log.Printf("Failed to unmarshal filemanager request: %v", err)
				msg.Nack(false, false)
				continue
			}
			response, err = h.handleDeleteFolder(request)
		default:
			err = fmt.Errorf("unknown queue: %s", queueName)
		}

		if err != nil {
			// Get transaction ID from response if available, otherwise use empty string
			transactionID := ""
			if response.TransactionID != "" {
				transactionID = response.TransactionID
			}
			response = messages.FileManagerResponse{
				TransactionID: transactionID,
				Success:       false,
				Error:         err.Error(),
			}
			log.Printf("Error handling request: %v", err)
		}

		// Send response (skip if empty - used for chunk intermediate responses)
		if response.TransactionID != "" {
			transactionID := response.TransactionID
			// For chunk responses, use "post.file" as the operation (not "post.file.chunk")
			// so the response routing key matches what the gateway is listening for
			responseQueueName := queueName
			if queueName == "filemanager.post.file.chunk" {
				responseQueueName = "filemanager.post.file"
			}
			if err := h.sendResponse(responseQueueName, transactionID, response, &msg); err != nil {
				log.Printf("Failed to send response: %v", err)
				continue
			}
		}

		// Acknowledge the message (chunks are already acknowledged in handleFileChunk)
		if queueName != "filemanager.post.file.chunk" {
			msg.Ack(false)
		}
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
	// This method is called with FileManagerRequest, but we need to check if the message
	// actually contains FileUploadRequest with file content
	// The actual handling is done in HandleFileManagerMessages where we unmarshal
	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       false,
		Error:         "PostFile requires file content, use handlePostFileWithContent instead",
	}, nil
}

// handleFileChunk handles a single file chunk and reassembles the file when all chunks are received
func (h *Handler) handleFileChunk(chunkRequest messages.FileChunkRequest, msg *amqp.Delivery) (messages.FileManagerResponse, error) {
	// Decode base64 chunk content
	chunkBytes, err := base64.StdEncoding.DecodeString(chunkRequest.Content)
	if err != nil {
		msg.Ack(false)
		return messages.FileManagerResponse{
			TransactionID: chunkRequest.TransactionID,
			Success:       false,
			Error:         fmt.Sprintf("failed to decode chunk %d: %v", chunkRequest.ChunkIndex, err),
		}, nil
	}

	// Store chunk
	h.chunkStorage.mu.Lock()

	// Initialize chunk map if needed
	if h.chunkStorage.chunks[chunkRequest.TransactionID] == nil {
		h.chunkStorage.chunks[chunkRequest.TransactionID] = make(map[int][]byte)
		h.chunkStorage.chunkCounts[chunkRequest.TransactionID] = chunkRequest.TotalChunks
		h.chunkStorage.metadata[chunkRequest.TransactionID] = &chunkMetadata{
			filename:  chunkRequest.Filename,
			storageID: chunkRequest.StorageID,
			totalSize: chunkRequest.TotalSize,
		}
	}

	// Store this chunk
	h.chunkStorage.chunks[chunkRequest.TransactionID][chunkRequest.ChunkIndex] = chunkBytes
	receivedChunks := len(h.chunkStorage.chunks[chunkRequest.TransactionID])
	totalChunks := h.chunkStorage.chunkCounts[chunkRequest.TransactionID]
	metadata := h.chunkStorage.metadata[chunkRequest.TransactionID]

	h.chunkStorage.mu.Unlock()

	// Acknowledge chunk message
	msg.Ack(false)

	// Check if we have all chunks
	if receivedChunks < totalChunks {
		// Still waiting for more chunks - don't send response yet
		log.Printf("Chunk %d/%d received for transaction %s, waiting for more", receivedChunks, totalChunks, chunkRequest.TransactionID)
		return messages.FileManagerResponse{}, nil // Empty response - don't send anything
	}

	// All chunks received - reassemble file
	log.Printf("All chunks received for transaction %s, reassembling file", chunkRequest.TransactionID)

	h.chunkStorage.mu.Lock()
	chunks := h.chunkStorage.chunks[chunkRequest.TransactionID]

	// Reassemble file in order
	reassembledFile := make([]byte, 0, chunkRequest.TotalSize)
	for i := 0; i < totalChunks; i++ {
		chunk, exists := chunks[i]
		if !exists {
			// Missing chunk - cleanup and return error
			delete(h.chunkStorage.chunks, chunkRequest.TransactionID)
			delete(h.chunkStorage.chunkCounts, chunkRequest.TransactionID)
			delete(h.chunkStorage.metadata, chunkRequest.TransactionID)
			h.chunkStorage.mu.Unlock()
			return messages.FileManagerResponse{
				TransactionID: chunkRequest.TransactionID,
				Success:       false,
				Error:         fmt.Sprintf("missing chunk %d", i),
			}, nil
		}
		reassembledFile = append(reassembledFile, chunk...)
	}

	// Cleanup chunk storage
	delete(h.chunkStorage.chunks, chunkRequest.TransactionID)
	delete(h.chunkStorage.chunkCounts, chunkRequest.TransactionID)
	delete(h.chunkStorage.metadata, chunkRequest.TransactionID)
	h.chunkStorage.mu.Unlock()

	// Process the reassembled file
	fileUpload := service.FileUpload{
		Filename: metadata.filename,
		Content:  bytes.NewReader(reassembledFile),
		Size:     chunkRequest.TotalSize,
	}

	// If storageID is provided, use PostFiles to add to existing storage
	// Otherwise, use PostFile to create new storage
	var result *service.UploadResult
	if metadata.storageID != "" {
		// Add file to existing storage
		result, err = h.service.PostFiles(h.ctx, chunkRequest.TransactionID, metadata.storageID, []service.FileUpload{fileUpload})
	} else {
		// Create new storage
		result, err = h.service.PostFile(h.ctx, chunkRequest.TransactionID, fileUpload)
	}

	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: chunkRequest.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}

	// Convert repository.FileInfo to messages.FileInfo
	files := make([]messages.FileInfo, len(result.Files))
	for i, fi := range result.Files {
		files[i] = messages.FileInfo{
			Filename: fi.Filename,
			Size:     fi.Size,
		}
	}

	return messages.FileManagerResponse{
		TransactionID: result.TransactionID,
		Success:       true,
		StorageID:     result.StorageID,
		Files:         files,
		TotalSize:     result.TotalSize,
	}, nil
}

// handlePostFileWithContent handles file upload with file content included in the message
func (h *Handler) handlePostFileWithContent(uploadRequest messages.FileUploadRequest) (messages.FileManagerResponse, error) {
	// Decode base64 content
	contentBytes, err := base64.StdEncoding.DecodeString(uploadRequest.Content)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: uploadRequest.TransactionID,
			Success:       false,
			Error:         fmt.Sprintf("failed to decode base64 content: %v", err),
		}, nil
	}

	// Create FileUpload struct
	fileUpload := service.FileUpload{
		Filename: uploadRequest.Filename,
		Content:  bytes.NewReader(contentBytes),
		Size:     uploadRequest.Size,
	}

	// If storageID is provided, use PostFiles to add to existing storage
	// Otherwise, use PostFile to create new storage
	var result *service.UploadResult
	if uploadRequest.StorageID != "" {
		// Add file to existing storage
		result, err = h.service.PostFiles(h.ctx, uploadRequest.TransactionID, uploadRequest.StorageID, []service.FileUpload{fileUpload})
	} else {
		// Create new storage
		result, err = h.service.PostFile(h.ctx, uploadRequest.TransactionID, fileUpload)
	}

	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: uploadRequest.TransactionID,
			Success:       false,
			Error:         err.Error(),
		}, nil
	}

	// Convert repository.FileInfo to messages.FileInfo
	files := make([]messages.FileInfo, len(result.Files))
	for i, fi := range result.Files {
		files[i] = messages.FileInfo{
			Filename: fi.Filename,
			Size:     fi.Size,
		}
	}

	return messages.FileManagerResponse{
		TransactionID: result.TransactionID,
		Success:       true,
		StorageID:     result.StorageID,
		Files:         files,
		TotalSize:     result.TotalSize,
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

	// TODO: optimize to stream for bigger files maybe through websockets
	// Read entire file into memory to get size and chunk it
	// For very large files, this could be optimized to stream, but for now this is simpler
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		return messages.FileManagerResponse{
			TransactionID: request.TransactionID,
			Success:       false,
			Error:         fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	fileSize := int64(len(fileContent))
	totalChunks := int((fileSize + ChunkSize - 1) / ChunkSize) // Ceiling division

	// Send file in chunks
	for chunkIndex := range totalChunks {
		start := chunkIndex * ChunkSize
		end := min(start+ChunkSize, len(fileContent))

		chunkData := fileContent[start:end]
		encodedChunk := base64.StdEncoding.EncodeToString(chunkData)

		chunkResponse := messages.FileChunkResponse{
			TransactionID: request.TransactionID,
			StorageID:     request.StorageID,
			Filename:      request.Filename,
			ChunkIndex:    chunkIndex,
			TotalChunks:   totalChunks,
			ChunkSize:     int64(len(chunkData)),
			TotalSize:     fileSize,
			Content:       encodedChunk,
			IsLastChunk:   chunkIndex == totalChunks-1,
		}

		chunkBody, err := json.Marshal(chunkResponse)
		if err != nil {
			log.Printf("Failed to marshal chunk %d: %v", chunkIndex, err)
			return messages.FileManagerResponse{
				TransactionID: request.TransactionID,
				Success:       false,
				Error:         fmt.Sprintf("failed to marshal chunk %d: %v", chunkIndex, err),
			}, nil
		}

		// Publish chunk with routing key: filemanager.response.get.file.chunk.<transaction-id>
		chunkRoutingKey := fmt.Sprintf("%s.%s", messages.TopicFileManagerGetFileChunk, request.TransactionID)
		if err := h.manager.PublishMessage(
			h.ctx,
			messages.FileManagerExchange,
			chunkRoutingKey,
			"application/json",
			chunkBody,
		); err != nil {
			log.Printf("Failed to publish chunk %d: %v", chunkIndex, err)
			return messages.FileManagerResponse{
				TransactionID: request.TransactionID,
				Success:       false,
				Error:         fmt.Sprintf("failed to publish chunk %d: %v", chunkIndex, err),
			}, nil
		}
	}

	// Return success response indicating file will be sent in chunks
	return messages.FileManagerResponse{
		TransactionID: request.TransactionID,
		Success:       true,
		StorageID:     request.StorageID,
		Data: map[string]interface{}{
			"filename":     request.Filename,
			"total_size":   fileSize,
			"total_chunks": totalChunks,
			"note":         "File content is being sent in chunks",
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
