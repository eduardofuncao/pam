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

type listFlags struct {
	oneline    bool
	searchTerm string
}

func parseListFlags() (listFlags, []string) {
	flags := listFlags{}
	remainingArgs := []string{}
	args := os.Args[2:]

	for _, arg := range args {
		if arg == "--oneline" || arg == "-o" {
			flags.oneline = true
		} else if !strings.HasPrefix(arg, "-") {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, remainingArgs
}

func (a *App) handleList() {
	flags, args := parseListFlags()

	var objectType string
	if len(args) == 0 {
		objectType = "queries" // Default to queries
	} else if args[0] == "queries" || args[0] == "connections" {
		objectType = args[0]
		if len(args) > 1 {
			flags.searchTerm = args[1]
		}
	} else {
		// First arg is search term, default to queries
		objectType = "queries"
		flags.searchTerm = args[0]
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

		// Get sorted list of queries
		queryList := make([]db.Query, 0, len(conn.Queries))
		for _, query := range conn.Queries {
			// If no search term, include all queries
			if flags.searchTerm == "" {
				queryList = append(queryList, query)
				continue
			}

			searchLower := strings.ToLower(flags.searchTerm)
			nameMatch := strings.Contains(strings.ToLower(query.Name), searchLower)
			sqlMatch := strings.Contains(strings.ToLower(query.SQL), searchLower)

			if nameMatch || sqlMatch {
				queryList = append(queryList, query)
			}
		}

		sort.Slice(queryList, func(i, j int) bool {
			return queryList[i].Id < queryList[j].Id
		})

		if flags.searchTerm != "" && len(queryList) == 0 {
			fmt.Printf(styles.Faint.Render("No queries found matching '%s'\n"), flags.searchTerm)
			return
		}

		// Display in oneline format if flag is set
		if flags.oneline {
			displayQueriesOneline(queryList)
			return
		}

		for _, query := range queryList {
			displayName := query.Name
			if flags.searchTerm != "" {
				displayName = highlightMatches(query.Name, flags.searchTerm)
			}

			// Extract table name
			tableName := db.ExtractTableNameFromSQL(query.SQL)
			if tableName == "" {
				tableName = "<unknown>"
			}
			if db.HasJoinClause(query.SQL) {
				tableName = tableName + " <join>"
			}

			formatedItem := fmt.Sprintf("◆ %d/%s (%s)", query.Id, displayName, tableName)
			fmt.Println(styles.Title.Render(formatedItem))

			displaySQL := query.SQL
			if flags.searchTerm != "" {
				displaySQL = highlightMatches(query.SQL, flags.searchTerm)
			}
			fmt.Print(parser.HighlightSQL(parser.FormatSQLWithLineBreaks(displaySQL)))
			fmt.Println()
			fmt.Println()
		}

	default:
		printError("Unknown list type: %s.  Use 'queries' or 'connections'", objectType)
	}
}

func displayQueriesOneline(queries []db.Query) {
	for _, query := range queries {
		tableName := db.ExtractTableNameFromSQL(query.SQL)
		hasJoin := db.HasJoinClause(query.SQL)

		tableDisplay := tableName
		if hasJoin && tableName != "" {
			tableDisplay = tableName + " <join>"
		} else if tableName == "" {
			tableDisplay = "<unknown>"
		}

		fmt.Printf("%s %s %s\n",
			styles.Faint.Render(fmt.Sprintf("%d", query.Id)),
			styles.Title.Render(query.Name),
			tableDisplay,
		)
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
