package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/initui"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleInit() {
	args := os.Args[1:]

	name, dbType, connString, schema := parseInitArgs(args)

	if dbType == "" && connString != "" {
		dbType = db.InferDBType(connString)
	}

	missing := false
	if name == "" {
		missing = true
	}
	if dbType == "" {
		missing = true
	}
	if connString == "" {
		missing = true
	}

	// Launch TUI
	if missing {
		var err error
		name, dbType, connString, err = initui.CollectInitParameters(name, dbType, connString)
		if err != nil {
			if err == initui.ErrAborted {
				fmt.Println(styles.Error.Render("✗ Init aborted"))
				os.Exit(0)
			}
			printError("Failed to collect parameters: %v", err)
		}
	}

	if name == "" || dbType == "" || connString == "" {
		printError("Missing required parameters: name, db-type, and connection-string are required")
	}

	conn, err := db.CreateConnection(name, dbType, connString)
	if err != nil {
		printError("Could not create connection interface: %s/%s, %s", dbType, name, err)
	}

	if schema != "" {
		conn.SetSchema(schema)
	}

	err = conn.Open()
	if err != nil {
		printError("Could not establish connection to: %s/%s: %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		printError("Could not communicate with the database: %s/%s, %s", dbType, name, err)
	}

	a.config.CurrentConnection = conn.GetName()
	a.config.Connections[a.config.CurrentConnection] = config.ToConnectionYAML(conn)
	err = a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	schemaInfo := ""
	if conn.GetSchema() != "" {
		schemaInfo = fmt.Sprintf(" (schema: %s)", conn.GetSchema())
	}
	fmt.Println(styles.Success.Render("✓ Connection created: "), styles.Title.Render(fmt.Sprintf("%s/%s%s", conn.GetDbType(), conn.GetName(), schemaInfo)))
}

// parseInitArgs parses flags and positional args
// Returns: name, dbType, connString, schema
func parseInitArgs(args []string) (string, string, string, string) {
	var name, dbType, connString, schema string

	// Check for flags
	hasFlags := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			hasFlags = true
			break
		}
	}

	if hasFlags {
		// Parse flag mode
		for i := 0; i < len(args); i++ {
			arg := args[i]

			switch {
			case arg == "--name" || arg == "-n":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					name = args[i+1]
					i++
				}
			case arg == "--type" || arg == "-t":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					dbType = args[i+1]
					i++
				}
			case arg == "--conn-string" || arg == "--conn" || arg == "-c":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					connString = args[i+1]
					i++
				}
			case arg == "--schema" || arg == "-s":
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					schema = args[i+1]
					i++
				}
			}
		}
	} else {
		// Positional mode:
		// 2-arg: name conn-string (type inferred)
		// 3-arg: name type conn-string [schema]
		if len(args) == 3 {
			// 2-arg mode: init name conn-string
			name = args[1]
			connString = args[2]
		} else if len(args) >= 4 {
			// 3-arg mode (legacy): init name type conn-string [schema]
			name = args[1]
			dbType = args[2]
			connString = args[3]
			if len(args) >= 5 {
				schema = args[4]
			}
		}
	}

	return name, dbType, connString, schema
}
