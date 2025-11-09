package repository

import (
	"context"
	"io"
)

type Repository interface {
	Close()
	SaveFile(ctx context.Context, sessionID string, filename string, content io.Reader) error
	GetFile(ctx context.Context, sessionID string, filename string) (io.ReadCloser, error)
	GetFilesBySession(ctx context.Context, sessionID string) ([]FileInfo, error)
	DeleteFile(ctx context.Context, sessionID string, filename string) error
	DeleteSession(ctx context.Context, sessionID string) error
}

// FileInfo represents metadata about a stored file
// Note: Path is intentionally omitted as it's implementation-specific.
// The repository abstraction should hide storage details (filesystem paths vs S3 keys).
type FileInfo struct {
	Filename string
	Size     int64
}
