package user

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalFileService(t *testing.T) {
	t.Run("creates service with valid parameters", func(t *testing.T) {
		service := NewLocalFileService("/tmp/test", "http://example.com", 5*1024*1024)
		
		assert.NotNil(t, service)
		assert.Equal(t, "/tmp/test", service.baseDir)
		assert.Equal(t, "http://example.com", service.baseURL)
		assert.Equal(t, int64(5*1024*1024), service.maxSize)
	})

	t.Run("uses default max size when zero", func(t *testing.T) {
		service := NewLocalFileService("/tmp/test", "http://example.com", 0)
		
		assert.Equal(t, int64(5*1024*1024), service.maxSize)
	})

	t.Run("uses default max size when negative", func(t *testing.T) {
		service := NewLocalFileService("/tmp/test", "http://example.com", -100)
		
		assert.Equal(t, int64(5*1024*1024), service.maxSize)
	})
}

func TestLocalFileService_SaveProfileImage(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	service := NewLocalFileService(tempDir, "http://example.com", 1024*1024)
	ctx := context.Background()

	t.Run("saves valid JPEG image", func(t *testing.T) {
		// Create minimal JPEG header to pass content type detection
		jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
		mimeType := "image/jpeg"
		userID := "user123"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, jpegData, mimeType)
		
		require.NoError(t, err)
		assert.Contains(t, url, "http://example.com")
		assert.NotEmpty(t, storagePath)
		assert.Contains(t, storagePath, userID)
		assert.Contains(t, storagePath, ".jpg")

		// Verify file was actually created
		fullPath := filepath.Join(tempDir, storagePath)
		assert.FileExists(t, fullPath)
	})

	t.Run("saves valid PNG image", func(t *testing.T) {
		// Create minimal PNG header
		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		mimeType := "image/png"
		userID := "user456"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, pngData, mimeType)
		
		require.NoError(t, err)
		assert.Contains(t, url, "http://example.com")
		assert.Contains(t, storagePath, ".png")

		// Verify file was actually created
		fullPath := filepath.Join(tempDir, storagePath)
		assert.FileExists(t, fullPath)
	})

	t.Run("rejects file that is too large", func(t *testing.T) {
		largeData := make([]byte, 2*1024*1024) // 2MB, larger than 1MB limit
		mimeType := "image/jpeg"
		userID := "user789"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, largeData, mimeType)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file too large")
		assert.Empty(t, url)
		assert.Empty(t, storagePath)
	})

	t.Run("rejects invalid MIME type", func(t *testing.T) {
		data := []byte("some data")
		mimeType := "text/plain"
		userID := "user999"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, data, mimeType)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported image type")
		assert.Empty(t, url)
		assert.Empty(t, storagePath)
	})

	t.Run("handles detected content type when mime not provided", func(t *testing.T) {
		// Text data that will be detected as text/plain
		textData := []byte("this is clearly not an image")
		mimeType := "" // Empty mime type - will detect content type
		userID := "user000"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, textData, mimeType)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported image type")
		assert.Empty(t, url)
		assert.Empty(t, storagePath)
	})

	t.Run("accepts empty data", func(t *testing.T) {
		data := []byte{}
		mimeType := "image/jpeg"
		userID := "user111"

		url, storagePath, err := service.SaveProfileImage(ctx, userID, data, mimeType)
		
		// The implementation actually allows empty data
		assert.NoError(t, err)
		assert.NotEmpty(t, url)
		assert.NotEmpty(t, storagePath)
	})
}

func TestLocalFileService_Delete(t *testing.T) {
	tempDir := t.TempDir()
	service := NewLocalFileService(tempDir, "http://example.com", 1024*1024)
	ctx := context.Background()

	t.Run("deletes existing file", func(t *testing.T) {
		// Create a test file
		testFile := "test/file.jpg"
		fullPath := filepath.Join(tempDir, testFile)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		
		err = os.WriteFile(fullPath, []byte("test content"), 0644)
		require.NoError(t, err)
		assert.FileExists(t, fullPath)

		// Delete the file
		err = service.Delete(ctx, testFile)
		
		assert.NoError(t, err)
		assert.NoFileExists(t, fullPath)
	})

	t.Run("handles non-existent file gracefully", func(t *testing.T) {
		err := service.Delete(ctx, "non/existent/file.jpg")
		
		// Should not return error for non-existent files
		assert.NoError(t, err)
	})

	t.Run("handles empty storage path", func(t *testing.T) {
		err := service.Delete(ctx, "")
		
		assert.NoError(t, err)
	})
}

func TestIsAllowedImageMime(t *testing.T) {
	testCases := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/jpg", true},
		{"image/png", true},
		{"IMAGE/JPEG", true}, // Case insensitive
		{"IMAGE/PNG", true},
		{"text/plain", false},
		{"application/pdf", false},
		{"image/gif", false}, // Not supported
		{"image/svg+xml", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.mimeType, func(t *testing.T) {
			result := isAllowedImageMime(tc.mimeType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtensionForMime(t *testing.T) {
	testCases := []struct {
		mimeType  string
		expected  string
	}{
		{"image/png", ".png"},
		{"IMAGE/PNG", ".png"}, // Case insensitive
		{"image/jpeg", ".jpg"},
		{"image/jpg", ".jpg"},
		{"text/plain", ".jpg"}, // Default to .jpg for unknown types
		{"", ".jpg"},
	}

	for _, tc := range testCases {
		t.Run(tc.mimeType, func(t *testing.T) {
			result := extensionForMime(tc.mimeType)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHttpDetectContentType(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "detects JPEG",
			data:     []byte{0xFF, 0xD8, 0xFF},
			expected: "image/jpeg",
		},
		{
			name:     "detects PNG",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expected: "image/png",
		},
		{
			name:     "detects text",
			data:     []byte("Hello World"),
			expected: "text/plain; charset=utf-8",
		},
		{
			name:     "handles empty data",
			data:     []byte{},
			expected: "text/plain; charset=utf-8",
		},
		{
			name:     "handles short data",
			data:     []byte{0xFF},
			expected: "text/plain; charset=utf-8",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := httpDetectContentType(tc.data)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMin(t *testing.T) {
	testCases := []struct {
		a, b     int
		expected int
	}{
		{5, 3, 3},
		{1, 10, 1},
		{0, 0, 0},
		{-5, -10, -10},
		{100, 50, 50},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := min(tc.a, tc.b)
			assert.Equal(t, tc.expected, result)
		})
	}
}