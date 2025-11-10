package repository

import (
	"context"
	"io"
)

type Repository interface {
	Close()
	SaveFile(ctx context.Context, storageID string, filename string, content io.Reader) error
	GetFile(ctx context.Context, storageID string, filename string) (io.ReadCloser, error)
	GetFilesByStorage(ctx context.Context, storageID string) ([]FileInfo, error)
	DeleteFile(ctx context.Context, storageID string, filename string) error
	DeleteStorage(ctx context.Context, storageID string) error
}

// FileInfo represents metadata about a stored file
// Note: Path is intentionally omitted as it's implementation-specific.
// The repository abstraction should hide storage details (filesystem paths vs S3 keys).
type FileInfo struct {
	Filename string
	Size     int64
}
