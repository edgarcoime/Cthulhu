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

// Dummy
// BookErrorResponse is the ErrorResponse that will be passed in the response by Handler
func BookErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}
