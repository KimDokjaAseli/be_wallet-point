package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	MaxUploadSize = 10 << 20 // 10MB
	UploadDir     = "../../public/uploads"
)

func HandleFileUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "No file uploaded", nil)
		return
	}

	if file.Size > MaxUploadSize {
		ErrorResponse(c, http.StatusBadRequest, "File size exceeds limit (10MB)", nil)
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), rand.Intn(1000), ext)

	// Ensure directory exists
	if _, err := os.Stat(UploadDir); os.IsNotExist(err) {
		os.MkdirAll(UploadDir, 0755)
	}

	savePath := filepath.Join(UploadDir, newFilename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err.Error())
		return
	}

	// Return relative URL for storage and retrieval
	fileURL := "/uploads/" + newFilename
	SuccessResponse(c, http.StatusOK, "File uploaded successfully", gin.H{
		"file_url": fileURL,
	})
}
