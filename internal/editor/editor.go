package editor

import (
	"fmt"
	"os"
	"os/exec"
)

func GetEditorCommand() string {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}
	return editorCmd
}

func EditTempFile(content, prefix string) (string, error) {
	tmpFile, err := CreateTempFile(prefix, content)
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	tmpFile.Close()

	editorCmd := GetEditorCommand()
	cmd := exec.Command(editorCmd, tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run editor: %w", err)
	}

	editedContent, err := ReadTempFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("read edited file: %w", err)
	}

	return editedContent, nil
}

func EditTempFileWithTemplate(template, prefix string) (string, error) {
	editedContent, err := EditTempFile(template, prefix)
	if err != nil {
		return "", err
	}

	// Strip instructions if present
	if HasInstructions(editedContent) {
		editedContent = StripInstructions(editedContent)
	}

	return editedContent, nil
}
