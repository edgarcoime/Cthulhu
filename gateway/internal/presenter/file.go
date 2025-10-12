package presenter

import (
	"github.com/gofiber/fiber/v2"
)

type File struct {
	OriginalName string `json:"original_name"`
	FileName     string `json:"file_name"`
	Size         int    `json:"size"`
	Path         string `json:"path"`
}

type FileInfo struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
}

func FileUploadSuccessResponse(url string, totalSize int, files *[]File) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data": fiber.Map{
			"url":        url,
			"files":      files,
			"total_size": totalSize,
			"file_count": len(*files),
		},
		"error": nil,
	}
}

func FileUploadErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   nil,
		"error":  err,
	}
}

func FileAccessSuccessResponse(sessionID string, files *[]FileInfo) *fiber.Map {
	return &fiber.Map{
		"status": true,
		"data": fiber.Map{
			"session_id": sessionID,
			"files":      files,
			"count":      len(*files),
		},
		"error": nil,
	}
}

func FileAccessErrorResponse(message string) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   nil,
		"error":  message,
	}
}

func FileDownloadErrorResponse(message string) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   nil,
		"error":  message,
	}
}

// Dummy
// BookErrorResponse is the ErrorResponse that will be passed in the response by Handler
func BookErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}
