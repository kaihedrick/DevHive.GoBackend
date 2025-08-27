package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FlyIOStorage provides file storage operations for Fly.io
type FlyIOStorage struct {
	BasePath string
}

// NewFlyIOStorage creates a new Fly.io storage instance
func NewFlyIOStorage() *FlyIOStorage {
	basePath := os.Getenv("FLY_VOLUME_PATH")
	if basePath == "" {
		basePath = "/app/static"
	}

	return &FlyIOStorage{
		BasePath: basePath,
	}
}

// SaveFile saves a file to Fly.io volume storage
func (f *FlyIOStorage) SaveFile(filename string, reader io.Reader) (string, error) {
	// Ensure the directory exists
	filePath := filepath.Join(f.BasePath, filename)
	dir := filepath.Dir(filePath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %v", err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy the content
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}

	// Return the public URL
	return fmt.Sprintf("/static/%s", filename), nil
}

// DeleteFile deletes a file from Fly.io volume storage
func (f *FlyIOStorage) DeleteFile(filename string) error {
	filePath := filepath.Join(f.BasePath, filename)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

// GetFileURL returns the public URL for a file
func (f *FlyIOStorage) GetFileURL(filename string) string {
	return fmt.Sprintf("/static/%s", filename)
}

// FileExists checks if a file exists
func (f *FlyIOStorage) FileExists(filename string) bool {
	filePath := filepath.Join(f.BasePath, filename)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetFileSize returns the size of a file in bytes
func (f *FlyIOStorage) GetFileSize(filename string) (int64, error) {
	filePath := filepath.Join(f.BasePath, filename)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %v", err)
	}

	return fileInfo.Size(), nil
}

// ListFiles lists all files in a directory
func (f *FlyIOStorage) ListFiles(dir string) ([]string, error) {
	dirPath := filepath.Join(f.BasePath, dir)

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	var fileList []string
	for _, file := range files {
		if !file.IsDir() {
			fileList = append(fileList, filepath.Join(dir, file.Name()))
		}
	}

	return fileList, nil
}

// CreateDirectory creates a directory if it doesn't exist
func (f *FlyIOStorage) CreateDirectory(dir string) error {
	dirPath := filepath.Join(f.BasePath, dir)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	return nil
}

// GetStorageInfo returns information about the storage
func (f *FlyIOStorage) GetStorageInfo() map[string]interface{} {
	// Get total size of storage
	var totalSize int64
	var fileCount int64

	err := filepath.Walk(f.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
			fileCount++
		}
		return nil
	})

	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to calculate storage info: %v", err),
		}
	}

	return map[string]interface{}{
		"base_path":  f.BasePath,
		"total_size": totalSize,
		"file_count": fileCount,
		"provider":   "fly.io",
	}
}
