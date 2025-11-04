package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edgarcoime/Cthulhu-gateway/internal/pkg"
	"github.com/edgarcoime/Cthulhu-gateway/internal/presenter"
	"github.com/gofiber/fiber/v2"
)

func UploadFile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		urlString := "testurl"
		var uploadedFiles []presenter.File
		var totalSize int64
		// timestamp := time.Now().Format("20060102_150405")

		res := presenter.FileUploadSuccessResponse(urlString, int(totalSize), &uploadedFiles)
		return c.JSON(res)
	}
}

func FileAccess() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var fileList []presenter.FileInfo

		res := presenter.FileAccessSuccessResponse(id, &fileList)
		return c.JSON(res)
	}
}

func FileDownload() fiber.Handler {
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
