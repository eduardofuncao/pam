package handler

import (
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/commands"
	"github.com/eduardofuncao/pam/internal/config"
)

func Parse(cfg *config.Config) {
	if len(os.Args) < 2 {
		commands.PrintHelp()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "help", "--help", "-h":
		commands.PrintHelp()
	case "init":
		commands.Init(cfg)
	case "switch", "use":
		commands.Switch(cfg)
	case "add", "save":
		commands.Add(cfg)
	case "remove", "delete":
		commands.Remove(cfg)
	case "run", "query":
		commands.Run(cfg)
	case "list", "ls":
		commands.List(cfg)
	case "edit":
		commands.Edit(cfg)
	case "status":
		commands.Status(cfg)
	case "history":
		commands.History(cfg)
	case "explore":
		commands.Explore(cfg)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

