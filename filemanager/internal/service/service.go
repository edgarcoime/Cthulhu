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
	SessionID string
	Files     []repository.FileInfo
	TotalSize int64
}

// Service interface defines the business logic layer for file management
type Service interface {
	// PostFile uploads a single file and creates a new session
	PostFile(ctx context.Context, file FileUpload) (*UploadResult, error)

	// PostFiles uploads multiple files to a session (creates new session if sessionID is empty)
	PostFiles(ctx context.Context, sessionID string, files []FileUpload) (*UploadResult, error)

	// GetFile retrieves a file by session ID and filename
	GetFile(ctx context.Context, sessionID string, filename string) (io.ReadCloser, error)

	// GetFiles retrieves all files in a session
	GetFiles(ctx context.Context, sessionID string) ([]repository.FileInfo, error)

	// DeleteFile deletes a specific file from a session
	DeleteFile(ctx context.Context, sessionID string, filename string) error

	// DeleteFolder deletes an entire session folder and all its files
	DeleteFolder(ctx context.Context, sessionID string) error
}
