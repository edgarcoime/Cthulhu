package repository

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type localRepository struct {
	// Db connection can be put here but since not needed leave open
	dirPath string
}

// NewLocalRepository creates a new local file repository instance
// dirPath is the base directory where all session folders will be stored
func NewLocalRepository(dirPath string) (*localRepository, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	r := &localRepository{
		dirPath: dirPath,
	}
	return r, nil
}

// Close implements the Repository interface
// For local repository, this is a no-op but kept for interface compliance
func (r *localRepository) Close() {
	// No cleanup needed for local file storage
}

// SaveFile saves a file to the session ID folder
// sessionID: 10-character UUID session identifier
// filename: name of the file to save
// content: reader containing the file content
func (r *localRepository) SaveFile(ctx context.Context, sessionID string, filename string, content io.Reader) error {
	// Validate session ID length (10 characters as per requirements)
	if len(sessionID) != 10 {
		return fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	// Create session directory path
	sessionDir := filepath.Join(r.dirPath, sessionID)

	// Create session directory if it doesn't exist
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create full file path
	filePath := filepath.Join(sessionDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content to file
	_, err = io.Copy(file, content)
	if err != nil {
		// Clean up the file if copy fails
		os.Remove(filePath)
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// GetFile retrieves a file by session ID and filename
// Returns a ReadCloser that must be closed by the caller
func (r *localRepository) GetFile(ctx context.Context, sessionID string, filename string) (io.ReadCloser, error) {
	// Validate session ID length
	if len(sessionID) != 10 {
		return nil, fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	// Create full file path
	filePath := filepath.Join(r.dirPath, sessionID, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filename)
	}

	// Open and return the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetFilesBySession retrieves all files in a session folder
func (r *localRepository) GetFilesBySession(ctx context.Context, sessionID string) ([]FileInfo, error) {
	// Validate session ID length
	if len(sessionID) != 10 {
		return nil, fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	// Create session directory path
	sessionDir := filepath.Join(r.dirPath, sessionID)

	// Check if session directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return []FileInfo{}, nil // Return empty slice if session doesn't exist
	}

	// Read directory contents
	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read session directory: %w", err)
	}

	var files []FileInfo
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Get file info
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't get info for
		}

		files = append(files, FileInfo{
			Filename: entry.Name(),
			Size:     info.Size(),
		})
	}

	return files, nil
}

// DeleteFile deletes a specific file from a session folder
func (r *localRepository) DeleteFile(ctx context.Context, sessionID string, filename string) error {
	// Validate session ID length
	if len(sessionID) != 10 {
		return fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	// Create full file path
	filePath := filepath.Join(r.dirPath, sessionID, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filename)
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteSession deletes an entire session folder and all its contents
func (r *localRepository) DeleteSession(ctx context.Context, sessionID string) error {
	// Validate session ID length
	if len(sessionID) != 10 {
		return fmt.Errorf("invalid session ID: must be exactly 10 characters")
	}

	// Create session directory path
	sessionDir := filepath.Join(r.dirPath, sessionID)

	// Check if session directory exists
	if _, err := os.Stat(sessionDir); os.IsNotExist(err) {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Remove the entire session directory
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("failed to delete session folder: %w", err)
	}

	return nil
}
