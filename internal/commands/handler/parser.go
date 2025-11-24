package handler

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardofuncao/pam/internal/commands"
	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
)

func Parse(cfg *config.Config) {
	ParseWithArgs(cfg, os.Args, false)
}

func ParseWithArgs(cfg *config.Config, args []string, fromTUI bool) (*db.TableData, error) {
	if len(args) < 2 {
		if fromTUI {
			return nil, fmt.Errorf("no command provided")
		}
		commands.PrintHelp("")
		os.Exit(1)
	}

	command := args[1]
	switch command {
	case "help", "--help", "-h":
		if fromTUI {
			return nil, fmt.Errorf("help command not available in TUI")
		}
		topic := ""
		if len(args) > 2 {
			topic = args[2]
		}
		commands.PrintHelp(topic)
	case "init":
		if fromTUI {
			return nil, fmt.Errorf("init command not available in TUI")
		}
		commands.Init(cfg)
	case "switch", "use":
		if fromTUI {
			return nil, fmt.Errorf("switch command not available in TUI")
		}
		commands.Switch(cfg)
	case "add", "save":
		if fromTUI {
			return nil, fmt.Errorf("add command not available in TUI")
		}
		commands.Add(cfg)
	case "remove", "delete":
		if fromTUI {
			return nil, fmt.Errorf("remove command not available in TUI")
		}
		commands.Remove(cfg)
	case "run", "query":
		cmdExec := func(args []string) (*db.TableData, error) {
			return ParseWithArgs(cfg, args, true)
		}
		return commands.RunWithArgs(cfg, args, fromTUI, cmdExec)
	case "list", "ls":
		if fromTUI {
			return nil, fmt.Errorf("list command not available in TUI")
		}
		commands.List(cfg)
	case "conf", "config":
		if fromTUI {
			return nil, fmt.Errorf("conf command not available in TUI")
		}
		commands.Edit(cfg)
	case "status":
		if fromTUI {
			return nil, fmt.Errorf("status command not available in TUI")
		}
		commands.Status(cfg)
	case "history":
		if fromTUI {
			return nil, fmt.Errorf("history command not available in TUI")
		}
		commands.History(cfg)
	case "explore":
		cmdExec := func(args []string) (*db.TableData, error) {
			return ParseWithArgs(cfg, args, true)
		}
		return commands.ExploreWithArgs(cfg, args, fromTUI, cmdExec)
	default:
		if fromTUI {
			return nil, fmt.Errorf("unknown command: %s", command)
		}
		log.Fatalf("Unknown command: %s", command)
	}

	return nil, nil
}

