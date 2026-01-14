package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/parser"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleList() {
	var objectType string
	if len(os.Args) < 3 {
		objectType = "queries" // Default to queries
	} else {
		objectType = os.Args[2]
	}

	switch objectType {
	case "connections":
		if len(a.config.Connections) == 0 {
			fmt.Println(styles.Faint.Render("No connections configured"))
			return
		}
		for name, connection := range a.config.Connections {
			marker := "◆"
			if name == a.config.CurrentConnection {
				marker = styles.Success.Render("●") // Active connection
			} else {
				marker = styles.Faint.Render("◆")
			}
			fmt.Printf("%s %s %s\n", marker, styles.Title.Render(name), styles.Faint.Render(fmt.Sprintf("(%s)", connection.DBType)))
		}

	case "queries":
		if a.config.CurrentConnection == "" {
			printError("No active connection.  Use 'pam switch <connection>' or 'pam init' first")
		}
		conn := a.config.Connections[a.config.CurrentConnection]
		if len(conn.Queries) == 0 {
			fmt.Println(styles.Faint.Render("No queries saved"))
			return
		}

		var searchTerm string
		if len(os.Args) > 3 {
			searchTerm = os.Args[3]
		}

		// Get sorted list of queries
		queryList := make([]db.Query, 0, len(conn.Queries))
		for _, query := range conn.Queries {
			// If no search term, include all queries
			if searchTerm == "" {
				queryList = append(queryList, query)
				continue
			}

			searchLower := strings.ToLower(searchTerm)
			nameMatch := strings.Contains(strings.ToLower(query.Name), searchLower)
			sqlMatch := strings.Contains(strings.ToLower(query.SQL), searchLower)

			if nameMatch || sqlMatch {
				queryList = append(queryList, query)
			}
		}

		sort.Slice(queryList, func(i, j int) bool {
			return queryList[i].Id < queryList[j].Id
		})

		if searchTerm != "" && len(queryList) == 0 {
			fmt.Printf(styles.Faint.Render("No queries found matching '%s'\n"), searchTerm)
			return
		}

		for _, query := range queryList {
			displayName := query.Name
			if searchTerm != "" {
				displayName = highlightMatches(query.Name, searchTerm)
			}
			formatedItem := fmt.Sprintf("◆ %d/%s", query.Id, displayName)
			fmt.Println(styles.Title.Render(formatedItem))

			displaySQL := query.SQL
			if searchTerm != "" {
				displaySQL = highlightMatches(query.SQL, searchTerm)
			}
			fmt.Print(parser.HighlightSQL(parser.FormatSQLWithLineBreaks(displaySQL)))
			fmt.Println()
		}

	default:
		printError("Unknown list type: %s.  Use 'queries' or 'connections'", objectType)
	}
}

func highlightMatches(text, searchTerm string) string {
	if searchTerm == "" {
		return text
	}

	searchLower := strings.ToLower(searchTerm)
	var result strings.Builder
	index := 0

	for {
		pos := strings.Index(strings.ToLower(text[index:]), searchLower)
		if pos == -1 {
			result.WriteString(text[index:])
			break
		}

		result.WriteString(text[index : index+pos])

		matchedText := text[index+pos : index+pos+len(searchTerm)]
		result.WriteString(styles.SearchMatch.Render(matchedText))

		index += pos + len(searchTerm)
	}

	return result.String()
}
