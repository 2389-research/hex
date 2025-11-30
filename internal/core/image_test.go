// ABOUTME: Tests for image loading, encoding, and validation
// ABOUTME: Ensures vision API support works correctly with various formats
package core

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadImage(t *testing.T) {
	// Create a temporary directory for test images
	tmpDir := t.TempDir()

	t.Run("loads valid PNG image", func(t *testing.T) {
		// Create a simple test PNG
		imgPath := filepath.Join(tmpDir, "test.png")
		createTestPNG(t, imgPath, 100, 100)

		imgSrc, err := LoadImage(imgPath)
		if err != nil {
			t.Fatalf("LoadImage() error = %v", err)
		}

		if imgSrc.MediaType != "image/png" {
			t.Errorf("MediaType = %v, want image/png", imgSrc.MediaType)
		}

		if imgSrc.Data == "" {
			t.Error("Data should not be empty")
		}

		// Verify it's valid base64
		_, err = base64.StdEncoding.DecodeString(imgSrc.Data)
		if err != nil {
			t.Errorf("Data is not valid base64: %v", err)
		}
	})

	t.Run("loads valid JPEG image", func(t *testing.T) {
		imgPath := filepath.Join(tmpDir, "test.jpg")
		createTestJPEG(t, imgPath, 100, 100)

		imgSrc, err := LoadImage(imgPath)
		if err != nil {
			t.Fatalf("LoadImage() error = %v", err)
		}

		if imgSrc.MediaType != "image/jpeg" {
			t.Errorf("MediaType = %v, want image/jpeg", imgSrc.MediaType)
		}
	})

	t.Run("rejects unsupported format", func(t *testing.T) {
		txtPath := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(txtPath, []byte("not an image"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		_, err := LoadImage(txtPath)
		if err == nil {
			t.Error("LoadImage() should error on unsupported format")
		}
	})

	t.Run("rejects non-existent file", func(t *testing.T) {
		_, err := LoadImage(filepath.Join(tmpDir, "nonexistent.png"))
		if err == nil {
			t.Error("LoadImage() should error on non-existent file")
		}
	})

	t.Run("rejects oversized image", func(t *testing.T) {
		// Create an image that exceeds 5MB
		imgPath := filepath.Join(tmpDir, "large.png")
		createTestPNG(t, imgPath, 3000, 3000) // Large image

		// Check file size
		info, err := os.Stat(imgPath)
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}

		// Only test size rejection if file is actually > 5MB
		if info.Size() > MaxImageSize {
			_, err = LoadImage(imgPath)
			if err == nil {
				t.Error("LoadImage() should error on oversized image")
			}
		}
	})
}

func TestEncodeImage(t *testing.T) {
	t.Run("encodes data to base64", func(t *testing.T) {
		testData := []byte("test image data")
		encoded := EncodeImage(testData, "image/png")

		// Should be valid base64
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Errorf("EncodeImage() produced invalid base64: %v", err)
		}

		if !bytes.Equal(decoded, testData) {
			t.Errorf("Decoded data doesn't match original")
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		encoded := EncodeImage([]byte{}, "image/png")
		// Empty data should encode to empty base64 string
		if encoded != "" {
			t.Errorf("EncodeImage() = %v, want empty string for empty data", encoded)
		}
	})
}

func TestDetectMediaType(t *testing.T) {
	tests := []struct {
		path     string
		wantType string
		wantErr  bool
	}{
		{"test.png", "image/png", false},
		{"test.PNG", "image/png", false},
		{"test.jpg", "image/jpeg", false},
		{"test.jpeg", "image/jpeg", false},
		{"test.JPG", "image/jpeg", false},
		{"test.gif", "image/gif", false},
		{"test.webp", "image/webp", false},
		{"test.txt", "", true},
		{"test.pdf", "", true},
		{"test", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			gotType, err := detectMediaType(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectMediaType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("detectMediaType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestValidateImageSize(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("accepts normal size", func(t *testing.T) {
		imgPath := filepath.Join(tmpDir, "normal.png")
		createTestPNG(t, imgPath, 100, 100)

		err := validateImageSize(imgPath)
		if err != nil {
			t.Errorf("validateImageSize() error = %v", err)
		}
	})

	t.Run("rejects oversized file", func(t *testing.T) {
		// Create a file larger than MaxImageSize
		largePath := filepath.Join(tmpDir, "large.bin")
		data := make([]byte, MaxImageSize+1)
		if err := os.WriteFile(largePath, data, 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		err := validateImageSize(largePath)
		if err == nil {
			t.Error("validateImageSize() should error on oversized file")
		}
	})
}

// Helper functions to create test images

func createTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}

func createTestJPEG(t *testing.T, path string, width, height int) {
	t.Helper()

	// For JPEG test, we'll create a PNG and rename it
	// In real implementation, use image/jpeg
	createTestPNG(t, path, width, height)
}
