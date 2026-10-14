package table

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// buildEditorCommand creates an exec. Cmd with cursor positioning based on the editor type
func buildEditorCommand(editorCmd, tmpPath, content string, cursorHint cursorPositionHint) *exec.Cmd {
	line, col := findCursorPosition(content, cursorHint)

	switch editorCmd {
	case "vim", "nvim":
		// For vim/neovim:  +call cursor(line, col)
		return exec.Command(editorCmd, fmt.Sprintf("+call cursor(%d,%d)", line, col), tmpPath)
	case "nano":
		// For nano: +LINE,COLUMN
		return exec.Command(editorCmd, fmt.Sprintf("+%d,%d", line, col), tmpPath)
	case "emacs":
		// For emacs: +LINE: COLUMN
		return exec.Command(editorCmd, fmt.Sprintf("+%d:%d", line, col), tmpPath)
	case "code", "vscode":
		// For VS Code: --goto file:line:column --wait
		return exec.Command(editorCmd, "--goto", fmt.Sprintf("%s:%d:%d", tmpPath, line, col), "--wait")
	default:
		// Fallback:  just open the file
		return exec.Command(editorCmd, tmpPath)
	}
}

type cursorPositionHint int

const (
	CursorAtTop         cursorPositionHint = iota // Top of file (line 1, col 1)
	CursorAtUpdateValue                           // Inside the value in UPDATE SET col = 'value'
	CursorAtWhereClause                           // Inside the value in WHERE col = 'value'
	CursorAtEndOfFile                             // At the end of the file
)

func findCursorPosition(content string, hint cursorPositionHint) (line int, col int) {
	lines := strings.Split(content, "\n")

	switch hint {
	case CursorAtTop:
		return 1, 1
	case CursorAtUpdateValue:
		// Look for:  SET column = 'value'
		re := regexp.MustCompile(`SET\s+\w+\s*=\s*'`)
		for i, lineText := range lines {
			match := re.FindStringIndex(lineText)
			if match != nil {
				return i + 1, match[1] + 1
			}
		}
		return 3, 1

	case CursorAtWhereClause:
		// Look for: WHERE column = 'value' and position inside the quotes
		re := regexp.MustCompile(`WHERE\s+\w+\s*=\s*'`)
		for i, lineText := range lines {
			match := re.FindStringIndex(lineText)
			if match != nil {
				return i + 1, match[1] + 1
			}
		}
		// Fallback: position after WHERE
		for i, lineText := range lines {
			if strings.Contains(strings.ToUpper(lineText), "WHERE") {
				idx := strings.Index(strings.ToUpper(lineText), "WHERE")
				if idx != -1 {
					return i + 1, idx + 7
				}
			}
		}
		return len(lines), 1

	case CursorAtEndOfFile:
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				return i + 1, len(lines[i]) + 1
			}
		}
		return len(lines), 1

	default:
		return 1, 1
	}
}
