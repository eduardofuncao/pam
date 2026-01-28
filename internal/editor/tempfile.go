package editor

import (
	"fmt"
	"os"
)

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

func ReadTempFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(data), nil
}
