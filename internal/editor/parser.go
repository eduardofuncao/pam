package editor

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	colorKeyword = "86"
	colorString  = "220"
)

var sqlKeywords = []string{
	"LEFT JOIN", "RIGHT JOIN", "INNER JOIN", "FULL JOIN", "CROSS JOIN",
	"FULL OUTER JOIN", "LEFT OUTER JOIN", "RIGHT OUTER JOIN",
	"INSERT INTO", "DELETE FROM", "GROUP BY", "ORDER BY",
	"UNION ALL", "FETCH FIRST",
	"SELECT", "FROM", "WHERE", "JOIN", "ON", "HAVING",
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

	for _, keyword := range sqlKeywords {
		pattern := regexp.MustCompile(`(?i)\s+` + regexp.QuoteMeta(keyword) + `\s+`)
		
		formatted = pattern.ReplaceAllStringFunc(formatted, func(match string) string {
			return "\n" + strings.TrimSpace(match) + " "
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
	keywordStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorKeyword)).
		Bold(true)

	stringStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorString))

	highlighted := sql

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
