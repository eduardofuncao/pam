package main

import (
	"fmt"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

func (a *App) handleInit() {
	if len(os.Args) < 5 {
		printError("Usage: pam init <name> <db-type> <connection-string>")
	}

	conn, err := db.CreateConnection(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		printError("Could not create connection interface: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	// ADD THIS:  Check if schema parameter is provided (5th argument)
	if len(os.Args) >= 6 && os.Args[5] != "" {
		conn.SetSchema(os.Args[5])
	}

	err = conn.Open()
	if err != nil {
		printError("Could not establish connection to:  %s/%s:  %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		printError("Could not communicate with the database: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	a.config. CurrentConnection = conn. GetName()
	a.config. Connections[a.config.CurrentConnection] = config.ToConnectionYAML(conn)
	err = a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	schemaInfo := ""
	if conn.GetSchema() != "" {
		schemaInfo = fmt.Sprintf(" (schema: %s)", conn.GetSchema())
	}
	fmt. Println(styles.Success.Render("✓ Connection created:  "), styles.Title.Render(fmt.Sprintf("%s/%s%s", conn.GetDbType(), conn.GetName(), schemaInfo)))
}
