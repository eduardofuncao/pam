package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/eduardofuncao/squix/internal/completion"
	"github.com/eduardofuncao/squix/internal/config"
)

func (a *App) handleComplete() {
	args := os.Args[2:]

	// Filter out empty strings (bash passes "" for current word being completed)
	var filteredArgs []string
	for _, arg := range args {
		if arg != "" {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	cfg, err := config.LoadConfig(config.CfgFile)
	if err != nil {
		return
	}

	completions := getCompletions(args, cfg)

	for _, c := range completions {
		fmt.Println(c)
	}
}

func (a *App) handleCompletion() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: squix completion bash|zsh|fish\n")
		os.Exit(1)
	}

	shell := os.Args[2]

	var script string
	switch shell {
	case "bash":
		script = completion.GenerateBash()
	case "zsh":
		script = completion.GenerateZsh()
	case "fish":
		script = completion.GenerateFish()
	default:
		fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
		os.Exit(1)
	}

	fmt.Print(script)
}

func getCompletions(args []string, cfg *config.Config) []string {
	if len(args) == 0 || (len(args) == 1 && args[0] == "") {
		return getSubcommands()
	}

	command := args[0]

	switch command {
	case "run", "query":
		// Check if completing after --format/-f
		for i, arg := range args {
			if (arg == "--format" || arg == "-f") && i == len(args)-1 {
				return []string{"csv", "json", "tsv", "html", "sql", "markdown"}
			}
		}
		result := getCurrentConnectionQueries(cfg)
		result = append(result, "--format", "-f")
		return result
	case "switch", "use":
		return getAllConnections(cfg)
	case "list", "ls":
		if len(args) >= 2 {
			if args[1] == "queries" {
				return getCurrentConnectionQueries(cfg)
			}
			if args[1] == "connections" {
				return []string{}
			}
			// Partial match for subcommands or query completion
			result := []string{"queries", "connections"}
			result = append(result, getCurrentConnectionQueries(cfg)...)
			return result
		}
		// No subcommand yet - return subcommands and query names
		result := []string{"queries", "connections"}
		result = append(result, getCurrentConnectionQueries(cfg)...)
		return result
	case "info":
		if len(args) >= 2 {
			return []string{"table", "view"}
		}
		return []string{"table", "view"}
	case "export":
		for i, arg := range args {
			if arg == "--format" && i == len(args)-1 {
				return []string{"custom", "plain", "tar"}
			}
		}
		if len(args) == 1 {
			return []string{"backup", "schema"}
		}
		result := []string{"--format"}
		if args[1] == "schema" {
			result = append(result, "--table")
		}
		return result
	case "edit", "delete", "rm", "remove":
		return getCurrentConnectionQueries(cfg)
	case "--connection", "-c":
		return getAllConnections(cfg)
	default:
		return getSubcommands()
	}
}

func getSubcommands() []string {
	return []string{
		"init",
		"switch",
		"use",
		"add",
		"save",
		"remove",
		"rm",
		"delete",
		"query",
		"run",
		"list",
		"ls",
		"edit",
		"info",
		"explore",
		"status",
		"test",
		"history",
		"tables",
		"t",
		"disconnect",
		"clear",
		"unset",
		"config",
		"explain",
		"example",
		"export",
		"help",
	}
}

func getAllConnections(cfg *config.Config) []string {
	var names []string
	for name := range cfg.Connections {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func getCurrentConnectionQueries(cfg *config.Config) []string {
	if cfg.CurrentConnection == "" {
		return []string{}
	}

	conn, exists := cfg.Connections[cfg.CurrentConnection]
	if !exists {
		return []string{}
	}

	var names []string
	for name := range conn.Queries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
