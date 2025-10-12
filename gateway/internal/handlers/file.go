package handlers

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
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

		return presenter.FileUploadSuccessResponse(urlString, int(totalSize), &uploadedFiles)
	}
}
