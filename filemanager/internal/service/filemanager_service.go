package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/edgarcoime/Cthulhu-filemanager/internal/repository"
	"github.com/google/uuid"
)

type fileManagerService struct {
	repository repository.Repository
}

// NewFileManagerService creates a new file manager service instance
func NewFileManagerService(r repository.Repository) Service {
	return &fileManagerService{
		repository: r,
	}
}

// generateStorageID creates a 10-character alphanumeric storage ID
// Uses Google's UUID package and extracts the first 10 alphanumeric characters
// The ID is URL-safe and case-sensitive (preserves hex case)
func generateStorageID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate storage ID: %w", err)
	}

	// Remove hyphens and take first 10 characters
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// After removing hyphens: 32 hex characters (0-9, a-f)
	// Taking first 10 gives us a 10-character alphanumeric string
	storageID := strings.ReplaceAll(id.String(), "-", "")
	if len(storageID) < 10 {
		return "", fmt.Errorf("generated UUID too short")
	}

	return storageID[:10], nil
}

// PostFile uploads a single file and creates a new storage location
func (s *fileManagerService) PostFile(ctx context.Context, transactionID string, file FileUpload) (*UploadResult, error) {
	// Validate transaction ID
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	// Generate new storage ID
	storageID, err := generateStorageID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate storage ID: %w", err)
	}

	// Save the file
	if err := s.repository.SaveFile(ctx, storageID, file.Filename, file.Content); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Get file info
	files, err := s.repository.GetFilesByStorage(ctx, storageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file info: %w", err)
	}

	return &UploadResult{
		TransactionID: transactionID,
		StorageID:     storageID,
		Files:         files,
		TotalSize:     file.Size,
	}, nil
}

// PostFiles uploads multiple files to a storage location
// If storageID is empty, a new storage location is created
func (s *fileManagerService) PostFiles(ctx context.Context, transactionID string, storageID string, files []FileUpload) (*UploadResult, error) {
	// Validate transaction ID
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Generate new storage ID if not provided
	if storageID == "" {
		var err error
		storageID, err = generateStorageID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate storage ID: %w", err)
		}
	} else if len(storageID) != 10 {
		return nil, fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	var totalSize int64
	// Save all files
	for _, file := range files {
		if err := s.repository.SaveFile(ctx, storageID, file.Filename, file.Content); err != nil {
			return nil, fmt.Errorf("failed to save file %s: %w", file.Filename, err)
		}
		totalSize += file.Size
	}

	// Get all file info
	fileInfos, err := s.repository.GetFilesByStorage(ctx, storageID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file info: %w", err)
	}

	return &UploadResult{
		TransactionID: transactionID,
		StorageID:     storageID,
		Files:         fileInfos,
		TotalSize:     totalSize,
	}, nil
}

// GetFile retrieves a file by storage ID and filename
func (s *fileManagerService) GetFile(ctx context.Context, transactionID string, storageID string, filename string) (io.ReadCloser, error) {
	// Validate transaction ID
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	if len(storageID) != 10 {
		return nil, fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	return s.repository.GetFile(ctx, storageID, filename)
}

// GetFiles retrieves all files in a storage location
func (s *fileManagerService) GetFiles(ctx context.Context, transactionID string, storageID string) ([]repository.FileInfo, error) {
	// Validate transaction ID
	if transactionID == "" {
		return nil, fmt.Errorf("transaction ID is required")
	}

	if len(storageID) != 10 {
		return nil, fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	return s.repository.GetFilesByStorage(ctx, storageID)
}

// DeleteFile deletes a specific file from a storage location
func (s *fileManagerService) DeleteFile(ctx context.Context, transactionID string, storageID string, filename string) error {
	// Validate transaction ID
	if transactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}

	if len(storageID) != 10 {
		return fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return s.repository.DeleteFile(ctx, storageID, filename)
}

// DeleteFolder deletes an entire storage folder and all its files
func (s *fileManagerService) DeleteFolder(ctx context.Context, transactionID string, storageID string) error {
	// Validate transaction ID
	if transactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}

	if len(storageID) != 10 {
		return fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	return s.repository.DeleteStorage(ctx, storageID)
}
