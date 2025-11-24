package commands

import (
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
)

func Init(cfg *config.Config) {
	if len(os.Args) < 5 {
		log.Fatal("Usage: pam create <name> <db-type> <connection-string> <user> <password>")
	}

	conn, err := db.CreateConnection(os.Args[2], os.Args[3], os.Args[4])
	if err != nil {
		log.Fatalf("Could not create connection interface: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	err = conn.Open()
	if err != nil {
		log.Fatalf("Could not establish connection to: %s/%s: %s",
			conn.GetDbType(), conn.GetName(), err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Fatalf("Could not communicate with the database: %s/%s, %s", os.Args[3], os.Args[2], err)
	}

	cfg.CurrentConnection = conn.GetName()
	cfg.Connections[cfg.CurrentConnection] = config.ToConnectionYAML(conn)
	cfg.Save()
}
