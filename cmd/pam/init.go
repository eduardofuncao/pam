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
		printError("Usage: pam create <name> <db-type> <connection-string> <user> <password>")
	}

	conn, err := db.CreateConnection(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		printError("Could not create connection interface: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	err = conn.Open()
	if err != nil {
		printError("Could not establish connection to:  %s/%s:  %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn. Ping()
	if err != nil {
		printError("Could not communicate with the database: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	a.config. CurrentConnection = conn.GetName()
	a.config. Connections[a.config.CurrentConnection] = config.ToConnectionYAML(conn)
	err = a.config.Save()
	if err != nil {
		printError("Could not save configuration file: %v", err)
	}

	fmt.Println(styles.Success.Render("✓ Connection created: "), styles.Title.Render(fmt.Sprintf("%s/%s", conn.GetDbType(), conn.GetName())))
}
