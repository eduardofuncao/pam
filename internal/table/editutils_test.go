package table

import (
	"strings"
	"testing"
)

func TestFindCursorPosition_UpdateValue(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantLine    int
		wantCol     int
		description string
	}{
		{
			name: "simple UPDATE with single quote",
			content: `UPDATE users
SET name = 'John Doe'
WHERE id = 1;`,
			wantLine:    2,
			wantCol:     13,
			description: "Should position cursor inside the value quotes in SET clause",
		},
		{
			name: "UPDATE with multiple spaces",
			content: `UPDATE employees
SET   salary   =   '50000'
WHERE employee_id = 123;`,
			wantLine:    2,
			wantCol:     20,
			description: "Should handle multiple spaces around SET clause",
		},
		{
			name: "UPDATE with mixed case",
			content: `update products
Set price = '99.99'
where product_id = 'P001';`,
			wantLine:    2,
			wantCol:     13,
			description: "Should work with lowercase SQL keywords",
		},
		{
			name: "no SET clause found",
			content: `SELECT * FROM users
WHERE id = 1;`,
			wantLine:    3,
			wantCol:     1,
			description: "Should fallback to default position when no SET clause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col := findCursorPosition(tt.content, CursorAtUpdateValue)
			if line != tt.wantLine {
				t.Errorf(
					"findCursorPosition() line = %v, want %v (%s)",
					line,
					tt.wantLine,
					tt.description,
				)
			}
			if col != tt.wantCol {
				t.Errorf(
					"findCursorPosition() col = %v, want %v (%s)",
					col,
					tt.wantCol,
					tt.description,
				)
			}
		})
	}
}

func TestFindCursorPosition_WhereClause(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantLine    int
		wantCol     int
		description string
	}{
		{
			name: "WHERE with single quote value",
			content: `DELETE FROM users
WHERE id = '123';`,
			wantLine:    2,
			wantCol:     12,
			description: "Should position cursor inside WHERE clause value",
		},
		{
			name: "WHERE with multiple spaces",
			content: `DELETE FROM employees
WHERE   employee_id   =   '456';`,
			wantLine:    2,
			wantCol:     27,
			description: "Should handle extra spaces in WHERE clause",
		},
		{
			name: "UPDATE with WHERE clause",
			content: `UPDATE users
SET status = 'active'
WHERE user_id = '789';`,
			wantLine:    3,
			wantCol:     17,
			description: "Should find WHERE clause even in UPDATE statements",
		},
		{
			name: "no quotes in WHERE - fallback to after WHERE keyword",
			content: `DELETE FROM users
WHERE id = 123;`,
			wantLine:    2,
			wantCol:     7,
			description: "Should position after WHERE when no quotes found",
		},
		{
			name:        "no WHERE clause at all",
			content:     `SELECT * FROM users;`,
			wantLine:    1,
			wantCol:     1,
			description: "Should fallback to end when no WHERE found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col := findCursorPosition(tt.content, CursorAtWhereClause)
			if line != tt.wantLine {
				t.Errorf(
					"findCursorPosition() line = %v, want %v (%s)",
					line,
					tt.wantLine,
					tt.description,
				)
			}
			if col != tt.wantCol {
				t.Errorf(
					"findCursorPosition() col = %v, want %v (%s)",
					col,
					tt.wantCol,
					tt.description,
				)
			}
		})
	}
}

func TestFindCursorPosition_EndOfFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantLine    int
		description string
	}{
		{
			name: "simple statement",
			content: `SELECT * FROM users
WHERE id = 1;`,
			wantLine:    2,
			description: "Should position at last non-empty line",
		},
		{
			name: "with trailing empty lines",
			content: `UPDATE users
SET name = 'test'
WHERE id = 1;


`,
			wantLine:    3,
			description: "Should skip trailing empty lines",
		},
		{
			name:        "single line",
			content:     `SELECT 1;`,
			wantLine:    1,
			description: "Should handle single line content",
		},
		{
			name: "only whitespace at end",
			content: `SELECT * FROM table;

	`,
			wantLine:    1,
			description: "Should skip whitespace-only lines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col := findCursorPosition(tt.content, CursorAtEndOfFile)
			if line != tt.wantLine {
				t.Errorf(
					"findCursorPosition() line = %v, want %v (%s)",
					line,
					tt.wantLine,
					tt.description,
				)
			}
			// Column should be at end of the line
			if col < 1 {
				t.Errorf(
					"findCursorPosition() col = %v, should be >= 1 (%s)",
					col,
					tt.description,
				)
			}
		})
	}
}

func TestBuildEditorCommand_Vim(t *testing.T) {
	content := `UPDATE users
SET name = 'test'
WHERE id = 1;`

	cmd := buildEditorCommand(
		"vim",
		"/tmp/test.sql",
		content,
		CursorAtUpdateValue,
	)

	if cmd.Path != "vim" && !strings.HasSuffix(cmd.Path, "/vim") {
		t.Errorf("Expected vim command, got %s", cmd.Path)
	}

	// Check that we have the cursor positioning argument
	found := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "+call cursor(") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected vim cursor positioning argument, got: %v", cmd.Args)
	}
}

func TestBuildEditorCommand_Neovim(t *testing.T) {
	content := `DELETE FROM users WHERE id = '123';`

	cmd := buildEditorCommand(
		"nvim",
		"/tmp/test.sql",
		content,
		CursorAtWhereClause,
	)

	if cmd.Path != "nvim" && !strings.HasSuffix(cmd.Path, "/nvim") {
		t.Errorf("Expected nvim command, got %s", cmd.Path)
	}

	// Check that we have the cursor positioning argument
	found := false
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "+call cursor(") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected nvim cursor positioning argument, got: %v", cmd.Args)
	}
}

func TestBuildEditorCommand_Nano(t *testing.T) {
	content := `SELECT * FROM users;`

	cmd := buildEditorCommand(
		"nano",
		"/tmp/test.sql",
		content,
		CursorAtEndOfFile,
	)

	if cmd.Path != "nano" && !strings.HasSuffix(cmd.Path, "/nano") {
		t.Errorf("Expected nano command, got %s", cmd.Path)
	}

	// Nano uses +LINE,COLUMN format
	found := false
	for _, arg := range cmd.Args {
		if strings.HasPrefix(arg, "+") && strings.Contains(arg, ",") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"Expected nano cursor positioning argument (+LINE,COL), got: %v",
			cmd.Args,
		)
	}
}

func TestBuildEditorCommand_Emacs(t *testing.T) {
	content := `UPDATE users SET name = 'test';`

	cmd := buildEditorCommand(
		"emacs",
		"/tmp/test.sql",
		content,
		CursorAtUpdateValue,
	)

	if cmd.Path != "emacs" && !strings.HasSuffix(cmd.Path, "/emacs") {
		t.Errorf("Expected emacs command, got %s", cmd.Path)
	}

	// Emacs uses +LINE:COLUMN format
	found := false
	for _, arg := range cmd.Args {
		if strings.HasPrefix(arg, "+") && strings.Contains(arg, ":") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"Expected emacs cursor positioning argument (+LINE:COL), got: %v",
			cmd.Args,
		)
	}
}

func TestBuildEditorCommand_VSCode(t *testing.T) {
	content := `SELECT * FROM users WHERE id = 1;`

	cmd := buildEditorCommand(
		"code",
		"/tmp/test.sql",
		content,
		CursorAtEndOfFile,
	)

	if cmd.Path != "code" && !strings.HasSuffix(cmd.Path, "/code") {
		t.Errorf("Expected code command, got %s", cmd.Path)
	}

	// VS Code uses --goto file:line:column --wait
	hasGoto := false
	hasWait := false
	for _, arg := range cmd.Args {
		if arg == "--goto" {
			hasGoto = true
		}
		if arg == "--wait" {
			hasWait = true
		}
	}
	if !hasGoto {
		t.Errorf("Expected --goto argument for VS Code, got: %v", cmd.Args)
	}
	if !hasWait {
		t.Errorf("Expected --wait argument for VS Code, got: %v", cmd.Args)
	}
}

func TestBuildEditorCommand_VSCodeAlias(t *testing.T) {
	content := `SELECT 1;`

	cmd := buildEditorCommand(
		"vscode",
		"/tmp/test.sql",
		content,
		CursorAtUpdateValue,
	)

	// Should work the same as "code"
	hasGoto := false
	hasWait := false
	for _, arg := range cmd.Args {
		if arg == "--goto" {
			hasGoto = true
		}
		if arg == "--wait" {
			hasWait = true
		}
	}
	if !hasGoto || !hasWait {
		t.Errorf(
			"Expected VS Code arguments for 'vscode' alias, got: %v",
			cmd.Args,
		)
	}
}

func TestBuildEditorCommand_UnknownEditor(t *testing.T) {
	content := `SELECT * FROM users;`

	cmd := buildEditorCommand(
		"unknown-editor",
		"/tmp/test.sql",
		content,
		CursorAtUpdateValue,
	)

	if cmd.Path != "unknown-editor" &&
		!strings.HasSuffix(cmd.Path, "/unknown-editor") {
		t.Errorf("Expected unknown-editor command, got %s", cmd.Path)
	}

	// Should just open the file without cursor positioning
	if len(cmd.Args) != 2 { // command name + file path
		t.Errorf(
			"Expected simple command with just file path, got: %v",
			cmd.Args,
		)
	}
}

func TestCursorPositionHint_Constants(t *testing.T) {
	// Verify that the constants are properly defined and distinct
	if CursorAtUpdateValue == CursorAtWhereClause {
		t.Error(
			"CursorAtUpdateValue should be different from CursorAtWhereClause",
		)
	}
	if CursorAtUpdateValue == CursorAtEndOfFile {
		t.Error(
			"CursorAtUpdateValue should be different from CursorAtEndOfFile",
		)
	}
	if CursorAtWhereClause == CursorAtEndOfFile {
		t.Error(
			"CursorAtWhereClause should be different from CursorAtEndOfFile",
		)
	}
}

func TestFindCursorPosition_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		hint    cursorPositionHint
		wantErr bool
	}{
		{
			name:    "empty content",
			content: "",
			hint:    CursorAtUpdateValue,
			wantErr: false, // Should handle gracefully with fallback
		},
		{
			name:    "only whitespace",
			content: "   \n\t\n  ",
			hint:    CursorAtEndOfFile,
			wantErr: false,
		},
		{
			name:    "very long line",
			content: strings.Repeat("A", 10000) + "\nSET col = 'value'",
			hint:    CursorAtUpdateValue,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, _ := findCursorPosition(tt.content, tt.hint)

			// Should always return valid positions
			if line < 1 {
				t.Errorf("Line should be >= 1, got %d", line)
			}
		})
	}
}

func TestFindCursorPosition_MultipleMatches(t *testing.T) {
	// When there are multiple matches, should find the first one
	content := `UPDATE users
SET name = 'first'
WHERE id = '123'
AND status = 'active';`

	line, _ := findCursorPosition(content, CursorAtUpdateValue)

	// Should find the first SET clause
	if line != 2 {
		t.Errorf("Expected to find first SET at line 2, got line %d", line)
	}
}

func TestFindCursorPosition_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name    string
		content string
		hint    cursorPositionHint
	}{
		{
			name:    "lowercase set",
			content: "update users\nset name = 'test'\nwhere id = 1;",
			hint:    CursorAtUpdateValue,
		},
		{
			name:    "uppercase SET",
			content: "UPDATE users\nSET name = 'test'\nWHERE id = 1;",
			hint:    CursorAtUpdateValue,
		},
		{
			name:    "mixed case SeT",
			content: "UpDaTe users\nSeT name = 'test'\nWhErE id = 1;",
			hint:    CursorAtUpdateValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, col := findCursorPosition(tt.content, tt.hint)

			// All should find the SET clause regardless of case
			if line != 2 {
				t.Errorf("%s: expected line 2, got %d", tt.name, line)
			}
			if col < 1 {
				t.Errorf("%s: expected valid column, got %d", tt.name, col)
			}
		})
	}
}
