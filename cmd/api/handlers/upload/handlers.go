package upload

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
)

// UploadFileRequest represents the file upload request
type UploadFileRequest struct {
	Filename    string `form:"filename" binding:"required"`
	FileType    string `form:"file_type" binding:"required"` // csv, parquet, jsonl
	Description string `form:"description"`
}

// UploadFileResponse represents the file upload response
type UploadFileResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message"`
	FilePath   string `json:"file_path"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	UploadTime string `json:"upload_time"`
	FileID     string `json:"file_id"`
}

// UploadedFile represents an uploaded file in the response
type UploadedFile struct {
	FileID     string `json:"file_id"`
	Filename   string `json:"filename"`
	FileSize   int64  `json:"file_size"`
	UploadTime string `json:"upload_time"`
	FileType   string `json:"file_type"`
	FilePath   string `json:"file_path"`
}

// UploadFile handles file uploads
func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get file from form
		file, err := c.FormFile("file")
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get file from form", err)
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "No file provided",
				Details: err.Error(),
			})
			return
		}

		// Get additional form data
		filename := c.PostForm("filename")

		if filename == "" {
			filename = file.Filename
		}

		// Validate file type
		allowedTypes := []string{"csv", "parquet", "jsonl", "json"}
		fileExt := strings.ToLower(filepath.Ext(filename))
		fileExt = strings.TrimPrefix(fileExt, ".")

		if !contains(allowedTypes, fileExt) {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Unsupported file type",
				Details: fmt.Sprintf("Supported types: %s", strings.Join(allowedTypes, ", ")),
			})
			return
		}

		// Create uploads directory if it doesn't exist
		uploadDir := "uploads"
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create uploads directory", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create upload directory",
				Details: err.Error(),
			})
			return
		}

		// Generate unique filename
		timestamp := time.Now().Format("20060102_150405")
		fileID := fmt.Sprintf("%s_%s", timestamp, filename)
		filePath := filepath.Join(uploadDir, fileID)

		// Save file
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			logger.LogError(logger.ServiceREST, "Failed to save uploaded file", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to save file",
				Details: err.Error(),
			})
			return
		}

		// Get file info
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get file info", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get file info",
				Details: err.Error(),
			})
			return
		}

		response := UploadFileResponse{
			Status:     "success",
			Message:    fmt.Sprintf("File uploaded successfully: %s", filename),
			FilePath:   filePath,
			Filename:   filename,
			FileSize:   fileInfo.Size(),
			UploadTime: time.Now().Format(time.RFC3339),
			FileID:     fileID,
		}

		logger.LogInfo(logger.ServiceREST, "File uploaded successfully", map[string]interface{}{
			"filename":  filename,
			"file_path": filePath,
			"file_size": fileInfo.Size(),
		})

		c.JSON(http.StatusOK, response)
	}
}

// ListUploadedFiles lists all uploaded files
func ListUploadedFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadDir := "uploads"

		// Check if uploads directory exists
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{
				"files": []interface{}{},
				"count": 0,
			})
			return
		}

		// Read directory
		files, err := os.ReadDir(uploadDir)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to read uploads directory", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to list files",
				Details: err.Error(),
			})
			return
		}

		// Build file list
		fileList := make([]UploadedFile, 0)
		for _, file := range files {
			if !file.IsDir() {
				fileInfo, err := file.Info()
				if err != nil {
					continue
				}

				fileList = append(fileList, UploadedFile{
					FileID:     file.Name(),
					Filename:   file.Name(),
					FileSize:   fileInfo.Size(),
					UploadTime: fileInfo.ModTime().Format(time.RFC3339),
					FileType:   strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Name()), ".")),
					FilePath:   filepath.Join(uploadDir, file.Name()),
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"files": fileList,
			"count": len(fileList),
		})
	}
}

// GetUploadedFile gets details of a specific uploaded file
func GetUploadedFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		if fileID == "" {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "File ID required",
				Details: "No file ID provided",
			})
			return
		}

		filePath := filepath.Join("uploads", fileID)

		// Check if file exists
		fileInfo, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error:   "File not found",
				Details: fmt.Sprintf("File %s does not exist", fileID),
			})
			return
		}
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get file info", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get file info",
				Details: err.Error(),
			})
			return
		}

		file := UploadedFile{
			FileID:     fileID,
			Filename:   fileInfo.Name(),
			FileSize:   fileInfo.Size(),
			UploadTime: fileInfo.ModTime().Format(time.RFC3339),
			FileType:   strings.ToLower(strings.TrimPrefix(filepath.Ext(fileInfo.Name()), ".")),
			FilePath:   filePath,
		}

		c.JSON(http.StatusOK, file)
	}
}

// DeleteUploadedFile deletes an uploaded file
func DeleteUploadedFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		if fileID == "" {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "File ID required",
				Details: "No file ID provided",
			})
			return
		}

		filePath := filepath.Join("uploads", fileID)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error:   "File not found",
				Details: fmt.Sprintf("File %s does not exist", fileID),
			})
			return
		}

		// Delete file
		if err := os.Remove(filePath); err != nil {
			logger.LogError(logger.ServiceREST, "Failed to delete file", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to delete file",
				Details: err.Error(),
			})
			return
		}

		logger.LogInfo(logger.ServiceREST, "File deleted successfully", map[string]interface{}{
			"file_id": fileID,
		})

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": fmt.Sprintf("File %s deleted successfully", fileID),
		})
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
