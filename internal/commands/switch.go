package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/config"
)

func Switch(cfg *config.Config) {
	if len(os.Args) < 3 {
		log.Fatal("Usage: pam switch/use <db-name>")
	}

	_, ok := cfg.Connections[os.Args[2]]
	if !ok {
		log.Fatalf("Connection %s does not exist", os.Args[2])
	}
	cfg.CurrentConnection = os.Args[2]

	err := cfg.Save()
	if err != nil {
		log.Fatal("Could not save configuration file")
	}
	fmt.Printf("connected to: %s/%s\n", cfg.Connections[cfg.CurrentConnection].DBType, cfg.CurrentConnection)
}
