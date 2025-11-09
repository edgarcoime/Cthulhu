package handlers

import (
	"fmt"
	"time"

	"github.com/edgarcoime/Cthulhu-common/pkg/messages"
	"github.com/edgarcoime/Cthulhu-gateway/internal/presenter"
	"github.com/edgarcoime/Cthulhu-gateway/internal/services"
	"github.com/gofiber/fiber/v2"
)

func RMQFileUpload(s *services.Container) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse multipart form to get all files
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(presenter.FileUploadErrorResponse(
				fmt.Errorf("failed to parse multipart form: %v. Please ensure you are sending a multipart/form-data request", err),
			))
		}

		// Get all files from the "file" field (supports multiple files)
		files := form.File["file"]
		if len(files) == 0 {
			return c.Status(400).JSON(presenter.FileUploadErrorResponse(
				fmt.Errorf("no files found. Please ensure you are sending files with field name 'file'"),
			))
		}

		// Get storage ID from query parameter (optional)
		storageID := c.Query("storage_id", "")

		var uploadedFiles []presenter.File
		var finalStorageID string
		uploadedFileNames := make(map[string]bool) // Track uploaded files to avoid duplicates
		var lastResponse *messages.FileManagerResponse

		// Upload files sequentially, reusing storageID from first file
		for i, file := range files {
			// Open the uploaded file
			fileHeader, err := file.Open()
			if err != nil {
				return c.Status(400).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("failed to open file %s: %w", file.Filename, err)))
			}

			// Get file size from the multipart file header
			fileSize := file.Size

			// Calculate timeout based on file size: 10 seconds per MB, minimum 30 seconds, maximum 5 minutes
			timeoutSeconds := int64(fileSize/(1024*1024))*10 + 30
			if timeoutSeconds < 30 {
				timeoutSeconds = 30
			}
			if timeoutSeconds > 300 {
				timeoutSeconds = 300 // 5 minutes max
			}
			timeout := time.Duration(timeoutSeconds) * time.Second

			// Use storageID from first file for subsequent files
			currentStorageID := storageID
			if i > 0 && finalStorageID != "" {
				currentStorageID = finalStorageID
			}

			// Upload file via RabbitMQ and wait for response
			response, err := s.FileHandler.UploadFileAndWait(
				file.Filename,
				fileHeader,
				fileSize,
				currentStorageID,
				timeout,
			)
			fileHeader.Close() // Close file after upload

			if err != nil {
				return c.Status(500).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("failed to upload file %s: %w", file.Filename, err)))
			}

			// Check if upload was successful
			if !response.Success {
				return c.Status(500).JSON(presenter.FileUploadErrorResponse(fmt.Errorf("file upload failed for %s: %s", file.Filename, response.Error)))
			}

			// Store storageID from first file to reuse for subsequent files
			if i == 0 {
				finalStorageID = response.StorageID
			}

			// Store last response to get final total size
			lastResponse = response

			// Aggregate file information (avoid duplicates since response includes all files in storage)
			for _, fileInfo := range response.Files {
				if !uploadedFileNames[fileInfo.Filename] {
					uploadedFileNames[fileInfo.Filename] = true
					uploadedFiles = append(uploadedFiles, presenter.File{
						OriginalName: fileInfo.Filename,
						FileName:     fileInfo.Filename,
						Size:         int(fileInfo.Size),
						Path:         fmt.Sprintf("/files/s/%s/d/%s", response.StorageID, fileInfo.Filename),
					})
				}
			}
		}

		// Generate URL
		urlString := fmt.Sprintf("/files/s/%s", finalStorageID)

		// Use total size from last response (includes all files in storage)
		totalSize := int64(0)
		if lastResponse != nil {
			totalSize = lastResponse.TotalSize
		}

		// Return success response
		res := presenter.FileUploadSuccessResponse(urlString, int(totalSize), &uploadedFiles)
		return c.JSON(res)
	}
}
