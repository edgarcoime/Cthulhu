package messages

// FileInfo represents metadata about a stored file
// This is a common representation used in messages
type FileInfo struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// FileManagerRequest represents a request for file operations
type FileManagerRequest struct {
	TransactionID string `json:"transaction_id"`
	StorageID     string `json:"storage_id,omitempty"` // Optional, used for operations on existing storage
	Filename      string `json:"filename,omitempty"`   // Optional, used for single file operations
	// For file uploads, the file content should be sent as a separate message or via a different mechanism
	// For now, we'll handle file content separately
}

// FileUploadRequest represents a file upload request with file content
// This extends FileManagerRequest to include the actual file content
type FileUploadRequest struct {
	TransactionID string `json:"transaction_id"`
	StorageID     string `json:"storage_id,omitempty"`
	Filename      string `json:"filename"`
	Content       string `json:"content"` // base64 encoded file content
	Size          int64  `json:"size"`
	IsChunked     bool   `json:"is_chunked,omitempty"` // true if this is part of a chunked upload
	ChunkIndex    int    `json:"chunk_index,omitempty"` // chunk index (0-based)
	TotalChunks   int    `json:"total_chunks,omitempty"` // total number of chunks
}

// FileChunkRequest represents a single chunk of a file
type FileChunkRequest struct {
	TransactionID string `json:"transaction_id"`
	StorageID     string `json:"storage_id,omitempty"`
	Filename      string `json:"filename"`
	ChunkIndex    int    `json:"chunk_index"`    // chunk index (0-based)
	TotalChunks   int    `json:"total_chunks"`   // total number of chunks
	ChunkSize    int64  `json:"chunk_size"`       // size of this chunk in bytes
	TotalSize    int64  `json:"total_size"`       // total file size in bytes
	Content      string `json:"content"`          // base64 encoded chunk content
}

// FileManagerResponse represents a response from filemanager service
type FileManagerResponse struct {
	TransactionID string                 `json:"transaction_id"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error,omitempty"`
	StorageID     string                 `json:"storage_id,omitempty"`
	Files         []FileInfo             `json:"files,omitempty"`
	TotalSize     int64                  `json:"total_size,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"` // For additional response data
}
