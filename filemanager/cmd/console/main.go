package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/edgarcoime/Cthulhu-filemanager/internal/repository"
	"github.com/edgarcoime/Cthulhu-filemanager/internal/service"
	"github.com/google/uuid"
)

const (
	defaultStorageDir = "/tmp/fileDump"
	usage             = `Usage: filemanager [OPTIONS]

Options:
  -u <filepath>          Upload a file or directory (creates new storage)
                        If path is a directory, uploads all files recursively
  -d                     Download a file
  -s <storage-id>        Storage ID (required for download)
  -f <filename>          Filename (required for download)
  -o <output-path>       Output path for download (default: current directory)
  -storage <dir>          Storage directory (default: /tmp/fileDump)

Examples:
  filemanager -u /path/to/file.txt
  filemanager -u /path/to/folder/
  filemanager -d -s abc123def4 -f file.txt
  filemanager -d -s abc123def4 -f file.txt -o /path/to/output/
`
)

func main() {
	var (
		uploadPath = flag.String("u", "", "Path to file to upload")
		download   = flag.Bool("d", false, "Download a file")
		storageID  = flag.String("s", "", "Storage ID (required for download)")
		filename   = flag.String("f", "", "Filename (required for download)")
		outputPath = flag.String("o", "", "Output path for download")
		storage    = flag.String("storage", defaultStorageDir, "Storage directory")
	)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Initialize repository and service
	repo, err := repository.NewLocalRepository(*storage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	fileService := service.NewFileManagerService(repo)
	ctx := context.Background()

	// Handle upload
	if *uploadPath != "" {
		if err := handleUpload(ctx, fileService, *uploadPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Upload failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Handle download
	if *download {
		if *storageID == "" {
			fmt.Fprintf(os.Stderr, "Error: Storage ID (-s) is required for download\n")
			os.Exit(1)
		}
		if *filename == "" {
			fmt.Fprintf(os.Stderr, "Error: Filename (-f) is required for download\n")
			os.Exit(1)
		}
		if err := handleDownload(ctx, fileService, *storageID, *filename, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Download failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No operation specified
	flag.Usage()
	os.Exit(1)
}

func handleUpload(ctx context.Context, fileService service.Service, filePath string) error {
	// Check if path is a directory or a file
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	if info.IsDir() {
		// Handle directory upload
		return handleDirectoryUpload(ctx, fileService, filePath)
	}

	// Handle single file upload
	return handleSingleFileUpload(ctx, fileService, filePath)
}

func handleSingleFileUpload(ctx context.Context, fileService service.Service, filePath string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Ensure file is at the beginning
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	// Extract filename from path
	filename := filepath.Base(filePath)

	// Generate transaction ID for this operation
	transactionID := uuid.New().String()

	// Create FileUpload struct
	fileUpload := service.FileUpload{
		Filename: filename,
		Content:  file,
		Size:     fileInfo.Size(),
	}

	// Upload the file
	result, err := fileService.PostFile(ctx, transactionID, fileUpload)
	if err != nil {
		return err
	}

	// Print success message
	fmt.Printf("✅ File uploaded successfully!\n")
	fmt.Printf("Transaction ID: %s\n", result.TransactionID)
	fmt.Printf("Storage ID: %s\n", result.StorageID)
	fmt.Printf("Filename: %s\n", result.Files[0].Filename)
	fmt.Printf("Size: %d bytes\n", result.Files[0].Size)
	fmt.Printf("Total size: %d bytes\n", result.TotalSize)
	fmt.Printf("\nTo download this file, use:\n")
	fmt.Printf("  filemanager -d -s %s -f %s\n", result.StorageID, filename)

	return nil
}

func handleDirectoryUpload(ctx context.Context, fileService service.Service, dirPath string) error {
	// First, collect all file paths
	var filePaths []struct {
		path     string
		relPath  string
		fileInfo os.FileInfo
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path for filename (preserves directory structure)
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			// If relative path fails, just use the base filename
			relPath = filepath.Base(path)
		}

		filePaths = append(filePaths, struct {
			path     string
			relPath  string
			fileInfo os.FileInfo
		}{path, relPath, info})

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(filePaths) == 0 {
		return fmt.Errorf("no files found in directory: %s", dirPath)
	}

	// Now open and upload all files
	var uploads []service.FileUpload
	for _, fp := range filePaths {
		file, err := os.Open(fp.path)
		if err != nil {
			// Close already opened files on error
			for _, upload := range uploads {
				if closer, ok := upload.Content.(io.Closer); ok {
					closer.Close()
				}
			}
			return fmt.Errorf("failed to open file %s: %w", fp.path, err)
		}

		// Ensure file is at the beginning
		if _, err := file.Seek(0, 0); err != nil {
			file.Close()
			// Close already opened files on error
			for _, upload := range uploads {
				if closer, ok := upload.Content.(io.Closer); ok {
					closer.Close()
				}
			}
			return fmt.Errorf("failed to seek file %s: %w", fp.path, err)
		}

		uploads = append(uploads, service.FileUpload{
			Filename: fp.relPath,
			Content:  file,
			Size:     fp.fileInfo.Size(),
		})
	}

	// Generate transaction ID for this operation
	transactionID := uuid.New().String()

	// Upload all files in a single storage location
	result, err := fileService.PostFiles(ctx, transactionID, "", uploads)
	if err != nil {
		// Close all open files on error
		for _, upload := range uploads {
			if closer, ok := upload.Content.(io.Closer); ok {
				closer.Close()
			}
		}
		return err
	}

	// Close all files after successful upload
	for _, upload := range uploads {
		if closer, ok := upload.Content.(io.Closer); ok {
			closer.Close()
		}
	}

	// Print success message
	fmt.Printf("✅ Successfully uploaded %d file(s)!\n", len(result.Files))
	fmt.Printf("Transaction ID: %s\n", result.TransactionID)
	fmt.Printf("Storage ID: %s\n", result.StorageID)
	fmt.Printf("Total size: %d bytes\n", result.TotalSize)
	fmt.Printf("\nUploaded files:\n")
	for _, file := range result.Files {
		fmt.Printf("  - %s (%d bytes)\n", file.Filename, file.Size)
	}
	fmt.Printf("\nTo download files, use:\n")
	for _, file := range result.Files {
		fmt.Printf("  filemanager -d -s %s -f %s\n", result.StorageID, file.Filename)
	}

	return nil
}

func handleDownload(ctx context.Context, fileService service.Service, storageID string, filename string, outputPath string) error {
	// Generate transaction ID for this operation
	transactionID := uuid.New().String()

	// Get the file
	fileReader, err := fileService.GetFile(ctx, transactionID, storageID, filename)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	// Determine the output path
	var finalOutputPath string
	if outputPath == "" {
		// Default to current directory with original filename
		finalOutputPath = filename
	} else {
		// Check if outputPath is a directory or a file path
		info, err := os.Stat(outputPath)
		if err == nil && info.IsDir() {
			// It's a directory, append the filename
			finalOutputPath = filepath.Join(outputPath, filename)
		} else if err != nil && os.IsNotExist(err) {
			// Path doesn't exist, check if it looks like a directory (ends with /)
			if len(outputPath) > 0 && (outputPath[len(outputPath)-1] == '/' || outputPath[len(outputPath)-1] == filepath.Separator) {
				// Create the directory and use filename
				if err := os.MkdirAll(outputPath, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
				finalOutputPath = filepath.Join(outputPath, filename)
			} else {
				// Treat as file path, create parent directory if needed
				parentDir := filepath.Dir(outputPath)
				if parentDir != "." && parentDir != "" {
					if err := os.MkdirAll(parentDir, 0755); err != nil {
						return fmt.Errorf("failed to create output directory: %w", err)
					}
				}
				finalOutputPath = outputPath
			}
		} else {
			// Some other error, treat as file path
			parentDir := filepath.Dir(outputPath)
			if parentDir != "." && parentDir != "" {
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
			}
			finalOutputPath = outputPath
		}
	}

	// Create output file
	outputFile, err := os.Create(finalOutputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Copy file content
	bytesWritten, err := io.Copy(outputFile, fileReader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Print success message
	fmt.Printf("✅ File downloaded successfully!\n")
	fmt.Printf("Transaction ID: %s\n", transactionID)
	fmt.Printf("Storage ID: %s\n", storageID)
	fmt.Printf("Filename: %s\n", filename)
	fmt.Printf("Size: %d bytes\n", bytesWritten)
	fmt.Printf("Saved to: %s\n", finalOutputPath)

	return nil
}
