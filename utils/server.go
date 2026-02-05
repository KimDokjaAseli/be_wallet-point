package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetServerURL constructs the base URL of the server from the request
func GetServerURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil || c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	// Check if host already contains scheme
	if strings.HasPrefix(c.Request.Host, "http://") || strings.HasPrefix(c.Request.Host, "https://") {
		return c.Request.Host
	}

	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}
