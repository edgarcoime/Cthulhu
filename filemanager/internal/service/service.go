package service

import (
	"context"
	"io"

	"github.com/edgarcoime/Cthulhu-filemanager/internal/repository"
)

// FileUpload represents a file to be uploaded
type FileUpload struct {
	Filename string
	Content  io.Reader
	Size     int64
}

// UploadResult represents the result of a file upload operation
type UploadResult struct {
	TransactionID string
	StorageID     string
	Files         []repository.FileInfo
	TotalSize     int64
}

// Service interface defines the business logic layer for file management
type Service interface {
	// PostFile uploads a single file and creates a new storage location
	// transactionID uniquely identifies this transaction in the saga pattern
	PostFile(ctx context.Context, transactionID string, file FileUpload) (*UploadResult, error)

	// PostFiles uploads multiple files to a storage location (creates new storage if storageID is empty)
	// transactionID uniquely identifies this transaction in the saga pattern
	PostFiles(ctx context.Context, transactionID string, storageID string, files []FileUpload) (*UploadResult, error)

	// GetFile retrieves a file by storage ID and filename
	// transactionID uniquely identifies this transaction in the saga pattern
	GetFile(ctx context.Context, transactionID string, storageID string, filename string) (io.ReadCloser, error)

	// GetFiles retrieves all files in a storage location
	// transactionID uniquely identifies this transaction in the saga pattern
	GetFiles(ctx context.Context, transactionID string, storageID string) ([]repository.FileInfo, error)

	// DeleteFile deletes a specific file from a storage location
	// transactionID uniquely identifies this transaction in the saga pattern
	DeleteFile(ctx context.Context, transactionID string, storageID string, filename string) error

	// DeleteFolder deletes an entire storage folder and all its files
	// transactionID uniquely identifies this transaction in the saga pattern
	DeleteFolder(ctx context.Context, transactionID string, storageID string) error
}
