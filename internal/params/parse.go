package params

import (
	"regexp"
	"strings"
)

var (
	// Matches :param_name|default_value
	// Captures: group 1 = param name, group 2 = default value
	// Supports quoted strings with spaces: 'Green apple'
	paramRegex = regexp.MustCompile(`:(\w+)(?:\|('(?:[^'\\]|\\.)*'|(?:[^'\s\\]+)))?`)
)

func ExtractParameters(sql string) map[string]string {
	params := make(map[string]string)

	// Remove SQL comments first to avoid false matches
	sqlWithoutComments := removeComments(sql)

	// Find all parameter definitions
	matches := paramRegex.FindAllStringSubmatch(sqlWithoutComments, -1)

	// Track seen params to handle duplicates (last one wins)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) >= 2 {
			paramName := match[1]
			defaultValue := ""
			if len(match) >= 3 && match[2] != "" {
				defaultValue = strings.TrimSpace(match[2])
				// Strip surrounding quotes for all databases
				// The database driver will add proper quoting
				if len(defaultValue) >= 2 && strings.HasPrefix(defaultValue, "'") && strings.HasSuffix(defaultValue, "'") {
					defaultValue = defaultValue[1 : len(defaultValue)-1]
					// Unescape escaped quotes
					defaultValue = strings.ReplaceAll(defaultValue, "''", "'")
					defaultValue = strings.ReplaceAll(defaultValue, "\\'", "'")
				}
			}

			// Only keep the last occurrence
			if !seen[paramName] {
				seen[paramName] = true
				params[paramName] = defaultValue
			}
		}
	}

	return params
}

// removeComments removes both -- and /* */ style comments from SQL
func removeComments(sql string) string {
	var result strings.Builder
	lines := strings.Split(sql, "\n")

	for _, line := range lines {
		// Remove -- comments
		if idx := strings.Index(line, "--"); idx != -1 {
			line = line[:idx]
		}
		result.WriteString(line + "\n")
	}

	// Now remove /* */ style comments from the whole text
	fullText := result.String()
	return removeBlockComments(fullText)
}

// removeBlockComments removes /* */ style comments
func removeBlockComments(sql string) string {
	var result strings.Builder
	inBlockComment := false

	for i := 0; i < len(sql); i++ {
		if i+1 < len(sql) && sql[i:i+2] == "/*" {
			inBlockComment = true
			i++ // skip next char
		} else if i+1 < len(sql) && sql[i:i+2] == "*/" {
			inBlockComment = false
			i++ // skip next char
		} else if !inBlockComment {
			result.WriteByte(sql[i])
		}
	}

	return result.String()
}
