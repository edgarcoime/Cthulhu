package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

const FILE_FOLDER = "/Users/edgarcoime/Documents/bscacs/cthulhu/code/Cthulhu/main/gateway/fileDump"

// generateRandomURL generates a unique random 10-character URL string (lowercase only)
func generateRandomURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 10

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	for {
		url := make([]byte, length)
		for i := range url {
			url[i] = charset[rand.Intn(len(charset))]
		}

		urlString := string(url)

		// Check if the generated URL already exists as a folder
		urlPath := filepath.Join(FILE_FOLDER, urlString)
		if _, err := os.Stat(urlPath); os.IsNotExist(err) {
			// Folder doesn't exist, this URL is unique
			return urlString
		}

		// Folder exists, generate another URL
	}
}

func main() {
	// Ensure fileDump directory exists
	if err := os.MkdirAll(FILE_FOLDER, 0755); err != nil {
		log.Fatalf("Failed to create fileDump directory: %v", err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
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

	// Routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})

	app.Post("/upload", handleFileUpload)
	app.Get("/files/:id", handleFileAccess)
	app.Get("/files/:id/download/:filename", handleFileDownload)

	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleFileUpload(c *fiber.Ctx) error {
	// Parse the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to parse multipart form",
		})
	}

	// Get the file from the form
	files := form.File["file"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// Get the first file (since we're handling just one file)
	file := files[0]

	// Generate a random URL string for this upload session
	urlString := generateRandomURL()

	// Create a folder for this upload session
	sessionFolder := filepath.Join(FILE_FOLDER, urlString)
	if err := os.MkdirAll(sessionFolder, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create session folder",
		})
	}

	// Generate a unique filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s%s", timestamp, file.Filename, ext)
	filePath := filepath.Join(sessionFolder, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to open uploaded file",
		})
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create destination file",
		})
	}
	defer dst.Close()

	// Copy the file content
	_, err = io.Copy(dst, src)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save file",
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"message":    "File uploaded successfully",
		"url":        urlString,
		"filename":   filename,
		"path":       filePath,
		"size":       file.Size,
		"session_id": urlString,
	})
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

			fileList = append(fileList, fiber.Map{
				"name": file.Name(),
				"size": fileInfo.Size(),
				"url":  fmt.Sprintf("/files/%s/download/%s", id, file.Name()),
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

	// Send the file
	return c.SendFile(filePath)
}
