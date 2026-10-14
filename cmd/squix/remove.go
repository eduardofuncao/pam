package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/squix/internal/config"
	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) handleRemove() {
	if len(os.Args) < 3 {
		printError("Usage: squix remove <run-name> [--connection|-c <conn-name>]")
	}

	// Parse flags
	var connectionName string
	args := os.Args[2:]
	for i, arg := range args {
		if arg == "--connection" || arg == "-c" {
			if i+1 < len(args) {
				connectionName = args[i+1]
			}
			break
		}
	}

	// If --connection flag was used, remove connection
	if connectionName != "" {
		a.removeConnection(connectionName)
		return
	}

	// Otherwise, remove query (original behavior)
	queries := a.config.QueriesFor(a.config.CurrentConnection)

	query, exists := db.FindQueryWithSelector(queries, os.Args[2])
	if !exists {
		printError("Query '%s' could not be found", os.Args[2])
	}

	delete(queries, query.Name)

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Removed run '%s'", query.Name)))
}

func (a *App) removeConnection(connName string) {
	if _, exists := a.config.Connections[connName]; !exists {
		printError("Connection '%s' does not exist", connName)
		return
	}

	key := config.GroupKey(connName)
	queryCount := len(a.config.QueriesFor(connName))
	hasSiblings := a.config.GroupHasOtherMembers(key, connName)

	if !a.confirmConnectionDeletion(connName, key, queryCount, hasSiblings) {
		fmt.Println(styles.Faint.Render("Aborted"))
		return
	}

	if a.config.CurrentConnection == connName {
		a.config.CurrentConnection = ""
	}

	delete(a.config.Connections, connName)
	// Drop the shared library only if no other connection still derives it.
	if !hasSiblings {
		delete(a.config.QueryGroups, key)
		// Remove the on-disk file too (only place group files are deleted).
		if err := os.Remove(config.GroupFile(key)); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "squix: could not remove %s: %v\n", config.GroupFile(key), err)
		}
	}

	err := a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
		return
	}

	if hasSiblings {
		fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Removed connection '%s' (group '%s' kept: %d queries shared with other connections)", connName, key, queryCount)))
	} else {
		fmt.Println(styles.Success.Render(fmt.Sprintf("✓ Removed connection '%s' and %d queries", connName, queryCount)))
	}
}

func (a *App) confirmConnectionDeletion(connName, groupKey string, queryCount int, hasSiblings bool) bool {
	reader := bufio.NewReader(os.Stdin)
	var prompt string
	if hasSiblings {
		prompt = fmt.Sprintf("This will delete connection '%s'. Its query group '%s' (%d queries) stays shared with other connections. Continue? [y/N]: ", connName, groupKey, queryCount)
	} else {
		prompt = fmt.Sprintf("This will delete connection '%s' and its %d queries. Continue? [y/N]: ", connName, queryCount)
	}
	fmt.Print(styles.Error.Render(prompt))

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
