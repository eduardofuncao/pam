package params

import (
	"strings"
)

func MapPositionalArgs(sql string, positionals []string) map[string]string {
	result := make(map[string]string)

	if len(positionals) == 0 {
		return result
	}

	// Extract parameters in order from SQL
	paramDefs := ExtractParameters(sql)

	// Get param names in order of appearance
	var paramNames []string
	seen := make(map[string]bool)

	// Parse SQL to find params in order
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		// Simple parsing: find :param patterns in order
		for i := 0; i < len(line); i++ {
			if line[i] == ':' && i+1 < len(line) {
				// Found potential param
				j := i + 1
				for j < len(line) && (line[j] >= 'a' && line[j] <= 'z' || line[j] >= 'A' && line[j] <= 'Z' || line[j] >= '0' && line[j] <= '9' || line[j] == '_') {
					j++
				}
				paramName := line[i+1 : j]

				// Check if this is a known parameter
				if _, exists := paramDefs[paramName]; exists && !seen[paramName] {
					paramNames = append(paramNames, paramName)
					seen[paramName] = true
				}
				i = j
			}
		}
	}

	// Map positionals to param names
	for i, value := range positionals {
		if i < len(paramNames) {
			result[paramNames[i]] = value
		}
	}

	return result
}
