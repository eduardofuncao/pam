package editor

import (
	"fmt"
	"os"
)

// CreateTempFile creates a temp file with the given prefix and writes content to it
// Returns the file handle (caller should close it), or error
func CreateTempFile(prefix, content string) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", prefix+"*.sql")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}

	return tmpFile, nil
}

// ReadTempFile reads content from a temp file path
// Returns the content as a string, or error
func ReadTempFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}
