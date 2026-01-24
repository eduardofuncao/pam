package editor

import (
	"fmt"
	"os"
	"os/exec"
)

// GetEditorCommand returns the editor command from EDITOR env var, defaults to "vim"
func GetEditorCommand() string {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vim"
	}
	return editorCmd
}

// EditTempFile opens an editor with the given content in a temp file
// Returns the edited content, or error if editor fails or content is empty
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

// EditTempFileWithTemplate opens an editor with a template in a temp file
// Returns the edited content with template instructions stripped
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
