package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-common/pkg/rabbitmq/manager"
	"github.com/google/uuid"
)

// ChunkSize is the maximum size of a chunk in bytes (1MB)
// This ensures messages stay well under RabbitMQ's practical limits
const ChunkSize = 1024 * 1024 // 1MB

// FileHandler handles file operations via RabbitMQ
type FileHandler struct {
	manager *manager.Manager
	ctx     context.Context
}

func NewFileHandler(rmqManager *manager.Manager, ctx context.Context) *FileHandler {
	return &FileHandler{
		manager: rmqManager,
		ctx:     ctx,
	}
}

func (h *FileHandler) SetupQueuesAndBindings() error {
	// Declare topic exchange for filemanager messages
	err := h.manager.DeclareExchange(
		messages.FileManagerExchange,
		"topic", // topic exchange type for flexible routing
		true,    // durable
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
	)
	if err != nil {
		return fmt.Errorf("failed to declare filemanager exchange: %w", err)
	}

	return nil
}

// UploadFile sends a file upload request to the filemanager service
// Automatically chunks large files to avoid RabbitMQ message size limits
// Streams chunks directly from the reader without loading entire file into memory
// Returns the transaction ID and any error
// NOTE: This method is kept for backward compatibility, but UploadFileAndWait should be used
// to avoid race conditions with response queue setup
func (h *FileHandler) UploadFile(filename string, fileContent io.Reader, fileSize int64, storageID string) (string, error) {
	// Generate transaction ID
	transactionID := uuid.New().String()
	
	err := h.UploadFileWithTransactionID(transactionID, filename, fileContent, fileSize, storageID)
	if err != nil {
		return "", err
	}
	
	return transactionID, nil
}

// uploadFileChunkedStreaming sends a file in chunks by streaming from the reader
// This avoids loading the entire file into memory
func (h *FileHandler) uploadFileChunkedStreaming(transactionID, filename string, fileContent io.Reader, storageID string, totalSize int64) (string, error) {
	// Calculate number of chunks
	totalChunks := int((totalSize + ChunkSize - 1) / ChunkSize) // Ceiling division

	// Use a buffered reader for efficient chunking
	buf := make([]byte, ChunkSize)
	chunkIndex := 0
	bytesRead := int64(0)

	// Stream and send chunks
	for bytesRead < totalSize {
		// Read one chunk at a time
		chunkSize := ChunkSize
		remaining := totalSize - bytesRead
		if remaining < ChunkSize {
			chunkSize = int(remaining)
		}

		n, err := io.ReadFull(fileContent, buf[:chunkSize])
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", fmt.Errorf("failed to read chunk %d: %w", chunkIndex, err)
		}
		if n == 0 {
			break
		}

		// Encode only the bytes we actually read
		chunkData := buf[:n]
		encodedChunk := base64.StdEncoding.EncodeToString(chunkData)

		chunkRequest := messages.FileChunkRequest{
			TransactionID: transactionID,
			StorageID:     storageID,
			Filename:      filename,
			ChunkIndex:    chunkIndex,
			TotalChunks:   totalChunks,
			ChunkSize:     int64(n),
			TotalSize:     totalSize,
			Content:       encodedChunk,
		}

		messageBody, err := json.Marshal(chunkRequest)
		if err != nil {
			return "", fmt.Errorf("failed to marshal chunk %d: %w", chunkIndex, err)
		}

		// Publish chunk
		if err := h.manager.PublishMessage(
			h.ctx,
			messages.FileManagerExchange,
			messages.TopicFileManagerPostFileChunk,
			"application/json",
			messageBody,
		); err != nil {
			return "", fmt.Errorf("failed to publish chunk %d: %w", chunkIndex, err)
		}

		bytesRead += int64(n)
		chunkIndex++

		// Small delay between chunks to avoid overwhelming RabbitMQ
		time.Sleep(10 * time.Millisecond)
	}

	return transactionID, nil
}

// WaitForResponse waits for a response from the filemanager service
// Returns the response and any error
func (h *FileHandler) WaitForResponse(transactionID string, timeout time.Duration) (*messages.FileManagerResponse, error) {
	// Create a temporary queue for receiving responses
	responseQueue, err := h.manager.DeclareQueue(
		"",    // let RabbitMQ generate a unique queue name
		false, // not durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare response queue: %w", err)
	}

	// Bind queue to receive responses for this transaction
	responseRoutingKey := fmt.Sprintf("%s.post.file.%s", messages.TopicFileManagerResponse, transactionID)
	if err := h.manager.QueueBind(
		responseQueue.Name,
		responseRoutingKey,
		messages.FileManagerExchange,
		false,
	); err != nil {
		return nil, fmt.Errorf("failed to bind response queue: %w", err)
	}

	// Consume messages from the response queue
	msgs, err := h.manager.Consume(
		responseQueue.Name,
		"",    // consumer tag (empty = auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	// Set up timeout
	ctx, cancel := context.WithTimeout(h.ctx, timeout)
	defer cancel()

	// Wait for response
	select {
	case msg := <-msgs:
		var response messages.FileManagerResponse
		if err := json.Unmarshal(msg.Body, &response); err != nil {
			msg.Nack(false, false)
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Acknowledge the message
		msg.Ack(false)

		return &response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for response: %w", ctx.Err())
	}
}

// UploadFileAndWait uploads a file and waits for the response
func (h *FileHandler) UploadFileAndWait(filename string, fileContent io.Reader, fileSize int64, storageID string, timeout time.Duration) (*messages.FileManagerResponse, error) {
	// Set up response queue BEFORE sending the file to avoid race conditions
	transactionID := uuid.New().String()
	
	// Create response queue first
	responseQueue, err := h.manager.DeclareQueue(
		"",    // let RabbitMQ generate a unique queue name
		false, // not durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare response queue: %w", err)
	}

	// Bind queue to receive responses for this transaction
	responseRoutingKey := fmt.Sprintf("%s.post.file.%s", messages.TopicFileManagerResponse, transactionID)
	if err := h.manager.QueueBind(
		responseQueue.Name,
		responseRoutingKey,
		messages.FileManagerExchange,
		false,
	); err != nil {
		return nil, fmt.Errorf("failed to bind response queue: %w", err)
	}

	// Start consuming BEFORE sending the request
	msgs, err := h.manager.Consume(
		responseQueue.Name,
		"",    // consumer tag (empty = auto-generated)
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	// Now send the file upload request
	err = h.UploadFileWithTransactionID(transactionID, filename, fileContent, fileSize, storageID)
	if err != nil {
		return nil, err
	}

	// Set up timeout
	ctx, cancel := context.WithTimeout(h.ctx, timeout)
	defer cancel()

	// Wait for response
	select {
	case msg := <-msgs:
		var response messages.FileManagerResponse
		if err := json.Unmarshal(msg.Body, &response); err != nil {
			msg.Nack(false, false)
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Acknowledge the message
		msg.Ack(false)

		return &response, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for response: %w", ctx.Err())
	}
}

// UploadFileWithTransactionID sends a file upload request with a pre-generated transaction ID
func (h *FileHandler) UploadFileWithTransactionID(transactionID, filename string, fileContent io.Reader, fileSize int64, storageID string) error {
	// Ensure exchange is declared (idempotent)
	if err := h.manager.DeclareExchange(
		messages.FileManagerExchange,
		"topic",
		true,
		false,
		false,
		false,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Determine if we need to chunk the file
	// Chunk if file is larger than ChunkSize
	if fileSize > ChunkSize {
		_, err := h.uploadFileChunkedStreaming(transactionID, filename, fileContent, storageID, fileSize)
		return err
	}

	// Small file - read and send as single message
	contentBytes, err := io.ReadAll(fileContent)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	encodedContent := base64.StdEncoding.EncodeToString(contentBytes)
	uploadRequest := messages.FileUploadRequest{
		TransactionID: transactionID,
		StorageID:     storageID,
		Filename:      filename,
		Content:       encodedContent,
		Size:          fileSize,
		IsChunked:     false,
	}

	messageBody, err := json.Marshal(uploadRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := h.manager.PublishMessage(
		h.ctx,
		messages.FileManagerExchange,
		messages.TopicFileManagerPostFile,
		"application/json",
		messageBody,
	); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
