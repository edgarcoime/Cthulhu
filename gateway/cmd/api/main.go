// Project naming conventions
// https://github.com/golang-standards/project-layout
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cthulhu-gateway/internal/pkg"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// Ensure fileDump directory exists
	if err := os.MkdirAll(pkg.FILE_FOLDER, 0o755); err != nil {
		log.Fatalf("Failed to create fileDump directory: %v", err)
	}

	app := fiber.New()

	// Add Cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: pkg.CORS_ORIGIN,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Middleware
	app.Use(logger.New(logger.Config{
		// 2005-03-19 15:10:26,618 - simple_example - DEBUG - debug mess
		Format:     "${date} ${time},${pid} - ${ip}:${port} - ${status} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
	}))
	app.Use(recover.New())

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	app.Post("/upload", handleFileUpload)
	app.Get("/files/:id", handleFileAccess)
	app.Get("/files/:id/download/:filename", handleFileDownload)

	if err := app.Listen(":" + pkg.PORT); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleFileAccess(c *fiber.Ctx) error {
	// Get the ID from the URL parameter
	id := c.Params("id")

	// Validate the ID format (10 characters, alphanumeric lowercase only)
	if len(id) != 10 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID format. Must be exactly 10 characters.",
		})
	}

	// Check if ID contains only valid characters
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid ID format. Only lowercase letters and numbers are allowed.",
			})
		}
	}

	// Check if the session folder exists
	sessionFolder := filepath.Join(FILE_FOLDER, id)
	if _, err := os.Stat(sessionFolder); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": "Session not found. The provided ID does not exist.",
		})
	}

	// Read all files in the session folder
	files, err := os.ReadDir(sessionFolder)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to read session files",
		})
	}

	// If no files found
	if len(files) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "No files found in this session",
		})
	}

	// Prepare file information
	var fileList []fiber.Map
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

			fileList = append(fileList, fiber.Map{
				"name":     originalName,
				"filename": file.Name(), // Keep the actual stored filename for download
				"size":     fileInfo.Size(),
				"url":      fmt.Sprintf("/files/%s/download/%s", id, file.Name()),
			})
		}
	}

	// Return the file list
	return c.JSON(fiber.Map{
		"session_id": id,
		"files":      fileList,
		"count":      len(fileList),
	})
}

func handleFileDownload(c *fiber.Ctx) error {
	// Get the ID and filename from URL parameters
	id := c.Params("id")
	filename := c.Params("filename")

	// Validate the ID format (same validation as handleFileAccess)
	if len(id) != 10 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ID format. Must be exactly 10 characters.",
		})
	}

	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid ID format. Only lowercase letters and numbers are allowed.",
			})
		}
	}

	// Check if the session folder exists
	sessionFolder := filepath.Join(FILE_FOLDER, id)
	if _, err := os.Stat(sessionFolder); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": "Session not found. The provided ID does not exist.",
		})
	}

	// Construct the full file path
	filePath := filepath.Join(sessionFolder, filename)

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": "File not found.",
		})
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
