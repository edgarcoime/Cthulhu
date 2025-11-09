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
