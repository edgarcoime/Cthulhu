package messages

// Topic constants for RabbitMQ routing
// Using hierarchical structure: diagnose.services.<operation>
const (
	// DiagnoseExchange is the exchange name for diagnostic messages
	DiagnoseExchange = "diagnose"

	// TopicDiagnoseServicesAll sends diagnostic messages to all services
	// Services should bind with: diagnose.services.*
	TopicDiagnoseServicesAll = "diagnose.services.all"

	// TopicDiagnoseServicesHealth is for health check requests
	TopicDiagnoseServicesHealth = "diagnose.services.health"

	// TopicDiagnoseServicesStatus is for service status requests
	TopicDiagnoseServicesStatus = "diagnose.services.status"

	// TopicDiagnoseServicesLoad is for service load/metrics requests
	TopicDiagnoseServicesLoad = "diagnose.services.load"

	// TopicDiagnoseServicesResponse is the topic for diagnostic responses (responses from services)
	// Format: diagnose.services.response.<service-name>
	TopicDiagnoseServicesResponse = "diagnose.services.response"
)

// FileManager topic constants for RabbitMQ routing
// Using hierarchical structure: filemanager.<operation>
const (
	// FileManagerExchange is the exchange name for file management messages
	FileManagerExchange = "filemanager"

	// Request topics - operations that filemanager service listens to
	// TopicFileManagerPostFile is for uploading a single file (creates new storage)
	TopicFileManagerPostFile = "filemanager.post.file"

	// TopicFileManagerPostFileChunk is for uploading a file chunk (part of chunked upload)
	TopicFileManagerPostFileChunk = "filemanager.post.file.chunk"

	// TopicFileManagerPostFiles is for uploading multiple files (creates new storage if storageID is empty)
	TopicFileManagerPostFiles = "filemanager.post.files"

	// TopicFileManagerGetFile is for retrieving a single file by storage ID and filename
	TopicFileManagerGetFile = "filemanager.get.file"

	// TopicFileManagerGetFiles is for retrieving all files in a storage location
	TopicFileManagerGetFiles = "filemanager.get.files"

	// TopicFileManagerDeleteFile is for deleting a specific file from a storage location
	TopicFileManagerDeleteFile = "filemanager.delete.file"

	// TopicFileManagerDeleteFolder is for deleting an entire storage folder and all its files
	TopicFileManagerDeleteFolder = "filemanager.delete.folder"

	// Response topics - responses from filemanager service
	// TopicFileManagerResponse is the base topic for responses
	// Format: filemanager.response.<operation>.<transaction-id>
	// Example: filemanager.response.post.file.abc123
	TopicFileManagerResponse = "filemanager.response"
)
