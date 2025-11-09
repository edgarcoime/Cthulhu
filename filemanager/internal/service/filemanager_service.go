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

// generateSessionID creates a 10-character alphanumeric session ID
// Uses Google's UUID package and extracts the first 10 alphanumeric characters
// The ID is URL-safe and case-sensitive (preserves hex case)
func generateSessionID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Remove hyphens and take first 10 characters
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	// After removing hyphens: 32 hex characters (0-9, a-f)
	// Taking first 10 gives us a 10-character alphanumeric string
	sessionID := strings.ReplaceAll(id.String(), "-", "")
	if len(sessionID) < 10 {
		return "", fmt.Errorf("generated UUID too short")
	}

	return sessionID[:10], nil
}

// PostFile uploads a single file and creates a new session
func (s *fileManagerService) PostFile(ctx context.Context, file FileUpload) (*UploadResult, error) {
	// Generate new session ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Save the file
	if err := s.repository.SaveFile(ctx, sessionID, file.Filename, file.Content); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Get file info
	files, err := s.repository.GetFilesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file info: %w", err)
	}

	return &UploadResult{
		SessionID: sessionID,
		Files:     files,
		TotalSize: file.Size,
	}, nil
}

// PostFiles uploads multiple files to a session
// If sessionID is empty, a new session is created
func (s *fileManagerService) PostFiles(ctx context.Context, sessionID string, files []FileUpload) (*UploadResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	// Generate new session ID if not provided
	if sessionID == "" {
		var err error
		sessionID, err = generateSessionID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session ID: %w", err)
		}
	} else if len(sessionID) != 10 {
		return nil, fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	var totalSize int64
	// Save all files
	for _, file := range files {
		if err := s.repository.SaveFile(ctx, sessionID, file.Filename, file.Content); err != nil {
			return nil, fmt.Errorf("failed to save file %s: %w", file.Filename, err)
		}
		totalSize += file.Size
	}

	// Get all file info
	fileInfos, err := s.repository.GetFilesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file info: %w", err)
	}

	return &UploadResult{
		SessionID: sessionID,
		Files:     fileInfos,
		TotalSize: totalSize,
	}, nil
}

// GetFile retrieves a file by session ID and filename
func (s *fileManagerService) GetFile(ctx context.Context, sessionID string, filename string) (io.ReadCloser, error) {
	if len(sessionID) != 10 {
		return nil, fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename cannot be empty")
	}

	return s.repository.GetFile(ctx, sessionID, filename)
}

// GetFiles retrieves all files in a session
func (s *fileManagerService) GetFiles(ctx context.Context, sessionID string) ([]repository.FileInfo, error) {
	if len(sessionID) != 10 {
		return nil, fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	return s.repository.GetFilesBySession(ctx, sessionID)
}

// DeleteFile deletes a specific file from a session
func (s *fileManagerService) DeleteFile(ctx context.Context, sessionID string, filename string) error {
	if len(sessionID) != 10 {
		return fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	return s.repository.DeleteFile(ctx, sessionID, filename)
}

// DeleteFolder deletes an entire session folder and all its files
func (s *fileManagerService) DeleteFolder(ctx context.Context, sessionID string) error {
	if len(sessionID) != 10 {
		return fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	return s.repository.DeleteSession(ctx, sessionID)
}
