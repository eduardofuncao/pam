package editor

import (
	"strings"
)

// HasInstructions checks if content contains instruction comments that should be stripped
func HasInstructions(content string) bool {
	return strings.Contains(content, "-- Enter your SQL") ||
		strings.Contains(content, "-- Enter your SQL run") ||
		strings.Contains(content, "-- Creating new") ||
		(strings.Contains(content, "--") && strings.Contains(content, "Save and exit"))
}

// StripInstructions removes instruction comments from content
// Looks for common instruction patterns and removes everything before the separator
func StripInstructions(content string) string {
	// Check for various instruction patterns
	patterns := []string{
		"-- Enter your SQL run below",
		"-- Enter your SQL query below",
		"-- Creating new",
	}

	// Find which pattern matches
	var separator string
	for _, pattern := range patterns {
		if strings.HasPrefix(strings.TrimSpace(content), pattern) ||
			strings.Contains(content, pattern) {
			separator = pattern
			break
		}
	}

	if separator == "" {
		// No recognized pattern, return as-is
		return strings.TrimSpace(content)
	}

	// Split and extract content after separator
	lines := strings.Split(content, "\n")
	var sqlLines []string
	foundSeparator := false
	foundDoubleDash := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for the separator line
		if strings.Contains(trimmed, separator) {
			foundSeparator = true
			continue
		}

		// After finding separator, look for "--" on its own line
		if foundSeparator && !foundDoubleDash {
			if trimmed == "--" {
				foundDoubleDash = true
				continue
			}
			// Skip lines between separator and "--"
			continue
		}

		// Collect SQL lines after "--" separator
		if foundDoubleDash {
			sqlLines = append(sqlLines, lines[i])
		}
	}

	if len(sqlLines) == 0 {
		// Fallback: if we didn't find the pattern, try returning everything after first empty line
		parts := strings.SplitN(content, "\n--\n", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
		return strings.TrimSpace(content)
	}

	return strings.TrimSpace(strings.Join(sqlLines, "\n"))
}
