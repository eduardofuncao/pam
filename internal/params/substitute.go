package params

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eduardofuncao/pam/internal/db"
)

func SubstituteParameters(sql string, paramValues map[string]string, conn db.DatabaseConnection) (string, []any, error) {
	if len(paramValues) == 0 {
		return sql, []any{}, nil
	}

	// Find all :param|default or :param patterns in order
	// Regex matches :param|value or :param where value can be quoted with spaces
	re := regexp.MustCompile(`:(\w+)(?:\|('(?:[^'\\]|\\.)*'|(?:[^'\s\\]+)))?`)
	matches := re.FindAllStringSubmatchIndex(sql, -1)

	if len(matches) == 0 {
		return sql, []any{}, nil
	}

	// Build ordered list of parameter values based on occurrence in SQL
	var orderedValues []any
	paramIndex := make(map[string]int) // Maps param name to its index (1-based)
	currentIndex := 1

	// Process matches in order of first appearance
	for _, match := range matches {
		// match[2:4] = group 1 (param name)
		paramName := sql[match[2]:match[3]]

		// Only add each param once (in order of first appearance)
		if _, exists := paramIndex[paramName]; !exists {
			if value, ok := paramValues[paramName]; ok {
				paramIndex[paramName] = currentIndex
				orderedValues = append(orderedValues, value)
				currentIndex++
			} else {
				return "", nil, fmt.Errorf("missing value for parameter: %s", paramName)
			}
		}
	}

	// Now replace :param|default or :param with appropriate placeholders
	result := replaceParamPlaceholders(sql, conn, paramIndex)

	return result, orderedValues, nil
}

// replaceParamPlaceholders replaces all :param|default or :param with DB-specific placeholders
func replaceParamPlaceholders(sql string, conn db.DatabaseConnection, paramIndex map[string]int) string {
	// Use same regex as initial extraction to handle quoted strings
	re := regexp.MustCompile(`:(\w+)(?:\|('(?:[^'\\]|\\.)*'|(?:[^'\s\\]+)))?`)

	// Replace all matches at once using ReplaceAllStringFunc
	result := re.ReplaceAllStringFunc(sql, func(match string) string {
		// Extract param name
		paramName := strings.TrimPrefix(match, ":")
		if pipeIdx := strings.Index(paramName, "|"); pipeIdx != -1 {
			paramName = paramName[:pipeIdx]
		}

		// Get placeholder for this param
		if index, ok := paramIndex[paramName]; ok {
			return conn.GetPlaceholder(index)
		}

		return match
	})

	return result
}

func GenerateDisplaySQL(sql string, paramValues map[string]string) string {
	// Find all :param|default or :param patterns
	re := regexp.MustCompile(`:(\w+)(?:\|('(?:[^'\\]|\\.)*'|(?:[^'\s\\]+)))?`)

	result := re.ReplaceAllStringFunc(sql, func(match string) string {
		// Extract param name
		paramName := strings.TrimPrefix(match, ":")
		if pipeIdx := strings.Index(paramName, "|"); pipeIdx != -1 {
			paramName = paramName[:pipeIdx]
		}

		// Get value for this param
		if value, ok := paramValues[paramName]; ok {
			// Try to determine if it's a number (unquoted) or string (quoted)
			// Simple heuristic: if it looks like a number, don't quote
			if isNumeric(value) {
				return value
			}
			// Quote strings and escape single quotes
			return "'" + strings.ReplaceAll(value, "'", "''") + "'"
		}

		return match
	})

	return result
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}

	// Check for integer or float format
	hasDigits := false
	hasDot := false

	for i, r := range s {
		if r >= '0' && r <= '9' {
			hasDigits = true
		} else if r == '.' && !hasDot && i > 0 && i < len(s)-1 {
			hasDot = true
		} else if (r == '-' || r == '+') && i == 0 {
			// Allow leading sign
		} else {
			return false
		}
	}

	return hasDigits
}
