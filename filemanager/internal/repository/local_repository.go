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

// SaveFile saves a file to the storage ID folder
// storageID: 10-character UUID storage identifier
// filename: name of the file to save
// content: reader containing the file content
func (r *localRepository) SaveFile(ctx context.Context, storageID string, filename string, content io.Reader) error {
	// Validate storage ID length (10 characters as per requirements)
	if len(storageID) != 10 {
		return fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	// Create storage directory path
	storageDir := filepath.Join(r.dirPath, storageID)

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create full file path
	filePath := filepath.Join(storageDir, filename)

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

// GetFile retrieves a file by storage ID and filename
// Returns a ReadCloser that must be closed by the caller
func (r *localRepository) GetFile(ctx context.Context, storageID string, filename string) (io.ReadCloser, error) {
	// Validate storage ID length
	if len(storageID) != 10 {
		return nil, fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	// Create full file path
	filePath := filepath.Join(r.dirPath, storageID, filename)

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

// GetFilesByStorage retrieves all files in a storage folder
func (r *localRepository) GetFilesByStorage(ctx context.Context, storageID string) ([]FileInfo, error) {
	// Validate storage ID length
	if len(storageID) != 10 {
		return nil, fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	// Create storage directory path
	storageDir := filepath.Join(r.dirPath, storageID)

	// Check if storage directory exists
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		return []FileInfo{}, nil // Return empty slice if storage doesn't exist
	}

	// Read directory contents
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
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

// DeleteFile deletes a specific file from a storage folder
func (r *localRepository) DeleteFile(ctx context.Context, storageID string, filename string) error {
	// Validate storage ID length
	if len(storageID) != 10 {
		return fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	// Create full file path
	filePath := filepath.Join(r.dirPath, storageID, filename)

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

// DeleteStorage deletes an entire storage folder and all its contents
func (r *localRepository) DeleteStorage(ctx context.Context, storageID string) error {
	// Validate storage ID length
	if len(storageID) != 10 {
		return fmt.Errorf("invalid storage ID: must be exactly 10 characters")
	}

	// Create storage directory path
	storageDir := filepath.Join(r.dirPath, storageID)

	// Check if storage directory exists
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		return fmt.Errorf("storage not found: %s", storageID)
	}

	// Remove the entire storage directory
	if err := os.RemoveAll(storageDir); err != nil {
		return fmt.Errorf("failed to delete storage folder: %w", err)
	}

	return nil
}
