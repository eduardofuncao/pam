package parser

import (
	"regexp"
	"strings"

	"github.com/eduardofuncao/pam/internal/styles"
)

var sqlKeywords = []string{
	"FULL OUTER JOIN", "LEFT OUTER JOIN", "RIGHT OUTER JOIN",
	"LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN",
	"INSERT INTO", "DELETE FROM", "GROUP BY", "ORDER BY",
	"UNION ALL", "FETCH FIRST",
	"SELECT", "FROM", "WHERE", "ON", "HAVING",
	"LIMIT", "OFFSET", "UNION", "UPDATE", "VALUES", "SET",
}

var highlightKeywords = []string{
	"SELECT", "FROM", "WHERE", "JOIN", "LEFT", "RIGHT", "INNER", "FULL", "CROSS", "OUTER",
	"ON", "GROUP", "BY", "HAVING", "ORDER", "LIMIT", "OFFSET", "UNION", "ALL",
	"INSERT", "INTO", "UPDATE", "DELETE", "VALUES", "SET", "AND", "OR", "NOT",
	"IN", "EXISTS", "BETWEEN", "LIKE", "IS", "NULL", "DISTINCT", "AS",
	"CASE", "WHEN", "THEN", "ELSE", "END", "FETCH", "FIRST", "ROWS", "ONLY",
}

func FormatSQLWithLineBreaks(sql string) string {
	if sql == "" {
		return ""
	}

	formatted := sql

	// Process keywords in order - longest first to avoid partial matches
	for _, keyword := range sqlKeywords {
		// Match keyword with word boundaries, case-insensitive
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)

		formatted = pattern.ReplaceAllStringFunc(formatted, func(match string) string {
			// Only add line break if not already at start of line
			if strings.HasPrefix(formatted[:strings.Index(formatted, match)], "\n") {
				return match + " "
			}
			return "\n" + match + " "
		})
	}

	lines := strings.Split(formatted, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}

	return strings.Join(cleanedLines, "\n")
}

func HighlightSQL(sql string) string {
	keywordStyle := styles.SQLKeyword
	stringStyle := styles.SQLString

	highlighted := sql

	// First, highlight compound keywords (multi-word)
	compoundKeywords := []string{
		"FULL OUTER JOIN", "LEFT OUTER JOIN", "RIGHT OUTER JOIN",
		"LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN",
		"INSERT INTO", "DELETE FROM", "GROUP BY", "ORDER BY",
		"UNION ALL", "FETCH FIRST", "ROWS ONLY",
	}

	for _, keyword := range compoundKeywords {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		highlighted = pattern.ReplaceAllStringFunc(highlighted, func(match string) string {
			return keywordStyle.Render(match)
		})
	}

	// Then, highlight individual keywords
	for _, keyword := range highlightKeywords {
		pattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
		highlighted = pattern.ReplaceAllStringFunc(highlighted, func(match string) string {
			return keywordStyle.Render(match)
		})
	}

	var result strings.Builder
	inString := false
	for _, char := range highlighted {
		if char == '\'' {
			if inString {
				result.WriteString(stringStyle.Render("'"))
				inString = false
			} else {
				result.WriteString(stringStyle.Render("'"))
				inString = true
			}
		} else if inString {
			result.WriteString(stringStyle.Render(string(char)))
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func countLines(s string) int {
	if s == "" {
		return 1
	}
	return strings.Count(s, "\n") + 1
}
