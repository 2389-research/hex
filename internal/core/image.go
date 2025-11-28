// ABOUTME: Image loading and encoding for vision API support
// ABOUTME: Handles file validation, size limits, and base64 encoding
package core

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// MaxImageSize is the maximum allowed image file size (5MB)
	// This matches Claude API limits
	MaxImageSize = 5 * 1024 * 1024
)

// ImageSource represents an image for vision API
type ImageSource struct {
	Type      string `json:"type"`       // Always "base64"
	MediaType string `json:"media_type"` // e.g., "image/png"
	Data      string `json:"data"`       // Base64 encoded image data
}

// LoadImage loads an image from a file path and returns an ImageSource
// Returns error if file doesn't exist, is too large, or has unsupported format
func LoadImage(path string) (*ImageSource, error) {
	// Check file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("image file not found: %w", err)
	}

	// Validate size
	if err := validateImageSize(path); err != nil {
		return nil, err
	}

	// Detect media type from extension
	mediaType, err := detectMediaType(path)
	if err != nil {
		return nil, err
	}

	// Read file data
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image file: %w", err)
	}

	// Encode to base64
	encoded := EncodeImage(data, mediaType)

	return &ImageSource{
		Type:      "base64",
		MediaType: mediaType,
		Data:      encoded,
	}, nil
}

// EncodeImage encodes image data to base64
func EncodeImage(data []byte, mediaType string) string {
	return base64.StdEncoding.EncodeToString(data)
}

// detectMediaType detects the media type from file extension
func detectMediaType(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".png":
		return "image/png", nil
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".gif":
		return "image/gif", nil
	case ".webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("unsupported image format: %s", ext)
	}
}

// validateImageSize checks if the image file is within size limits
func validateImageSize(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Size() > MaxImageSize {
		return fmt.Errorf("image file too large: %d bytes (max %d bytes)", info.Size(), MaxImageSize)
	}

	return nil
}
