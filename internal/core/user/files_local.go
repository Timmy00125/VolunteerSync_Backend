package user

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LocalFileService stores files on local filesystem under baseDir and serves via baseURL.
type LocalFileService struct {
	baseDir string
	baseURL string
	maxSize int64 // bytes
}

// NewLocalFileService constructs a LocalFileService.
func NewLocalFileService(baseDir, baseURL string, maxSize int64) *LocalFileService {
	if maxSize <= 0 {
		maxSize = 5 * 1024 * 1024
	}
	return &LocalFileService{baseDir: baseDir, baseURL: baseURL, maxSize: maxSize}
}

func (l *LocalFileService) SaveProfileImage(ctx context.Context, userID string, data []byte, mimeType string) (string, string, error) {
	if int64(len(data)) > l.maxSize {
		return "", "", fmt.Errorf("file too large")
	}
	// Validate mime
	if mimeType == "" {
		mt := httpDetectContentType(data)
		mimeType = mt
	}
	if !isAllowedImageMime(mimeType) {
		return "", "", fmt.Errorf("unsupported image type: %s", mimeType)
	}

	// Generate deterministic name
	sum := sha1.Sum(append([]byte(userID), data[:min(len(data), 1024)]...))
	name := hex.EncodeToString(sum[:])
	ext := extensionForMime(mimeType)
	relPath := filepath.Join("profiles", userID, fmt.Sprintf("%s%s", name, ext))
	absPath := filepath.Join(l.baseDir, relPath)

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", "", fmt.Errorf("mkdir: %w", err)
	}

	// Save file bytes as-is
	if err := os.WriteFile(absPath, data, 0o644); err != nil {
		return "", "", fmt.Errorf("write: %w", err)
	}

	url := strings.TrimRight(l.baseURL, "/") + "/" + filepath.ToSlash(relPath)
	return url, relPath, nil
}

func (l *LocalFileService) Delete(ctx context.Context, storagePath string) error {
	if storagePath == "" {
		return nil
	}
	abs := filepath.Join(l.baseDir, storagePath)
	if err := os.Remove(abs); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func isAllowedImageMime(mt string) bool {
	switch strings.ToLower(mt) {
	case "image/jpeg", "image/jpg", "image/png":
		return true
	default:
		return false
	}
}

func extensionForMime(mt string) string {
	switch strings.ToLower(mt) {
	case "image/png":
		return ".png"
	default:
		return ".jpg"
	}
}

// httpDetectContentType wraps http.DetectContentType without importing net/http globally here.
func httpDetectContentType(b []byte) string {
	// Minimal 512 bytes for detection
	n := 512
	if len(b) < n {
		n = len(b)
	}
	return http.DetectContentType(b[:n])
}

// min helper
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
