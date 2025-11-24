package commands

import (
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
)

func Remove(cfg *config.Config) {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam remove <query-name>")
	}

	conn := cfg.Connections[cfg.CurrentConnection]
	queries := conn.Queries

	query, exists := db.FindQueryWithSelector(queries, os.Args[2])
	if exists {
		delete(conn.Queries, query.Name)
	} else {
		log.Fatalf("Query %s could not be found", os.Args[2])
	}
	err := cfg.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}
}
