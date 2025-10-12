package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cthulhu-gateway/internal/pkg"
	"cthulhu-gateway/internal/presenter"
	"cthulhu-gateway/pkg/file"

	"github.com/gofiber/fiber/v2"
)

// generateRandomURL generates a unique random 10-character URL string (lowercase only)
func generateRandomURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 10

	// Seed the random number generator
	for {
		url := make([]byte, length)
		for i := range url {
			url[i] = charset[rand.Intn(len(charset))]
		}

		urlString := string(url)

		// Check if the generated URL already exists as a folder
		urlPath := filepath.Join(pkg.FILE_FOLDER, urlString)
		if _, err := os.Stat(urlPath); os.IsNotExist(err) {
			// Folder doesn't exist, this URL is unique
			return urlString
		}

		// Folder exists, generate another URL
	}
}

func UploadFile(service file.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse the multipart form
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Failed to parse multipart form",
			})
		}

		// Get the files from the form
		files := form.File["file"]
		if len(files) == 0 {
			return c.Status(400).JSON(fiber.Map{
				"error": "No files provided",
			})
		}

		// Generate a random URL string for this upload session
		urlString := generateRandomURL()

		// Create a folder for this upload session
		sessionFolder := filepath.Join(pkg.FILE_FOLDER, urlString)
		if err := os.MkdirAll(sessionFolder, 0o755); err != nil {
			return c.Status(500).JSON(presenter.FileUploadErrorResponse(err))
		}

		var uploadedFiles []presenter.File
		var totalSize int64
		timestamp := time.Now().Format("20060102_150405")

		for i, file := range files {

			// Generate a unique filename with timestamp and index
			// Keep the original filename but add timestamp and index prefix
			filename := fmt.Sprintf("%s_%d_%s", timestamp, i+1, file.Filename)
			filePath := filepath.Join(sessionFolder, filename)

			// Open the uploaded file
			src, err := file.Open()
			if err != nil {
				return c.Status(500).JSON(presenter.FileUploadErrorResponse(err))
			}

			// Create the destination file
			dst, err := os.Create(filePath)
			if err != nil {
				src.Close()
				return c.Status(500).JSON(presenter.FileUploadErrorResponse(err))
			}

			// Copy the file content
			_, err = io.Copy(dst, src)
			src.Close()
			dst.Close()

			if err != nil {
				return c.Status(500).JSON(presenter.FileUploadErrorResponse(err))
			}

			// Add file info to the response
			newFile := presenter.File{
				OriginalName: file.Filename,
				FileName:     filename,
				Size:         int(file.Size),
				Path:         filePath,
			}
			uploadedFiles = append(uploadedFiles, newFile)

			totalSize += file.Size
		}

		return c.JSON(presenter.FileUploadSuccessResponse(urlString, int(totalSize), &uploadedFiles))
	}
}

func FileAccess(service file.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the ID from the URL parameter
		id := c.Params("id")

		// Validate the ID format (10 characters, alphanumeric lowercase only)
		if len(id) != 10 {
			return c.Status(400).JSON(presenter.FileAccessErrorResponse("Invalid ID format. Must be exactly 10 characters."))
		}

		// Check if ID contains only valid characters
		for _, char := range id {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				return c.Status(400).JSON(presenter.FileAccessErrorResponse("Invalid ID format. Only lowercase letters and numbers are allowed."))
			}
		}

		// Check if the session folder exists
		sessionFolder := filepath.Join(pkg.FILE_FOLDER, id)
		if _, err := os.Stat(sessionFolder); os.IsNotExist(err) {
			return c.Status(404).JSON(presenter.FileAccessErrorResponse("Session not found. The provided ID does not exist."))
		}

		// Read all files in the session folder
		files, err := os.ReadDir(sessionFolder)
		if err != nil {
			return c.Status(500).JSON(presenter.FileAccessErrorResponse("Failed to read session files"))
		}

		// If no files found
		if len(files) == 0 {
			return c.Status(404).JSON(presenter.FileAccessErrorResponse("No files found in this session"))
		}

		// Prepare file information
		var fileList []presenter.FileInfo
		for _, file := range files {
			if !file.IsDir() {
				filePath := filepath.Join(sessionFolder, file.Name())
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					continue // Skip files that can't be read
				}

				// Extract original filename by removing timestamp and index prefix
				originalName := file.Name()
				// Remove timestamp_XX_ prefix to get original filename
				if parts := strings.SplitN(file.Name(), "_", 3); len(parts) >= 3 {
					originalName = parts[2] // Get everything after timestamp_XX_
				}

				fileList = append(fileList, presenter.FileInfo{
					Name:     originalName,
					Filename: file.Name(), // Keep the actual stored filename for download
					Size:     fileInfo.Size(),
					URL:      fmt.Sprintf("/files/%s/download/%s", id, file.Name()),
				})
			}
		}

		// Return the file list
		return c.JSON(presenter.FileAccessSuccessResponse(id, &fileList))
	}
}

func FileDownload(service file.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the ID and filename from URL parameters
		id := c.Params("id")
		filename := c.Params("filename")

		// Validate the ID format (same validation as FileAccess)
		if len(id) != 10 {
			return c.Status(400).JSON(presenter.FileDownloadErrorResponse("Invalid ID format. Must be exactly 10 characters."))
		}

		for _, char := range id {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				return c.Status(400).JSON(presenter.FileDownloadErrorResponse("Invalid ID format. Only lowercase letters and numbers are allowed."))
			}
		}

		// Check if the session folder exists
		sessionFolder := filepath.Join(pkg.FILE_FOLDER, id)
		if _, err := os.Stat(sessionFolder); os.IsNotExist(err) {
			return c.Status(404).JSON(presenter.FileDownloadErrorResponse("Session not found. The provided ID does not exist."))
		}

		// Construct the full file path
		filePath := filepath.Join(sessionFolder, filename)

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(404).JSON(presenter.FileDownloadErrorResponse("File not found."))
		}

		// Extract original filename for download
		originalName := filename
		if parts := strings.SplitN(filename, "_", 3); len(parts) >= 3 {
			originalName = parts[2] // Get everything after timestamp_XX_
		}

		// Set the Content-Disposition header to use the original filename
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalName))

		// Send the file
		return c.SendFile(filePath)
	}
}
