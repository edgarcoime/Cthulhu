package handlers

import (
	"fmt"
	"time"

	"github.com/edgarcoime/Cthulhu-gateway/internal/presenter"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

func RMQFileUpload(s *services.Container) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the file from the request
		// FormFile will return an error if:
		// 1. Content-Type is not multipart/form-data
		// 2. No file field named "file" is present
		// 3. The request body is malformed
		file, err := c.FormFile("file")
		if err != nil {
			// Provide a more helpful error message
			return c.Status(400).JSON(presenter.FileUploadErrorResponse(
				fmt.Errorf("failed to get file: %v. Please ensure you are sending a multipart/form-data request with a file field named 'file'", err),
			))
		}

		// Open the uploaded file
		fileHeader, err := file.Open()
		if err != nil {
			return c.Status(400).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("failed to open file: %w", err)))
		}
		defer fileHeader.Close()

		// Get file size from the multipart file header (no need to read entire file)
		fileSize := file.Size

		// Get storage ID from query parameter (optional)
		storageID := c.Query("storage_id", "")

		// Upload file via RabbitMQ and wait for response
		// Calculate timeout based on file size: 10 seconds per MB, minimum 30 seconds, maximum 5 minutes
		timeoutSeconds := int64(fileSize/(1024*1024))*10 + 30
		if timeoutSeconds < 30 {
			timeoutSeconds = 30
		}
		if timeoutSeconds > 300 {
			timeoutSeconds = 300 // 5 minutes max
		}
		timeout := time.Duration(timeoutSeconds) * time.Second
		
		// Pass the file reader directly - it will stream chunks without loading entire file into memory
		response, err := s.FileHandler.UploadFileAndWait(
			file.Filename,
			fileHeader,
			fileSize,
			storageID,
			timeout,
		)
		if err != nil {
			return c.Status(500).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("failed to upload file: %w", err)))
		}

		// Check if upload was successful
		if !response.Success {
			return c.Status(500).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("file upload failed: %s", response.Error)))
		}

		// Convert response to presenter format
		var uploadedFiles []presenter.File
		for _, fileInfo := range response.Files {
			uploadedFiles = append(uploadedFiles, presenter.File{
				OriginalName: fileInfo.Filename,
				FileName:     fileInfo.Filename,
				Size:         int(fileInfo.Size),
				Path:         fmt.Sprintf("/files/s/%s/d/%s", response.StorageID, fileInfo.Filename),
			})
		}

		// Generate URL (you may want to adjust this based on your routing)
		urlString := fmt.Sprintf("/files/s/%s", response.StorageID)

		// Return success response
		res := presenter.FileUploadSuccessResponse(urlString, int(response.TotalSize), &uploadedFiles)
		return c.JSON(res)
	}
}
