package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eduardofuncao/squix/internal/explain"
)

type explainFlags struct {
	depth   int
	verbose bool
}

func parseExplainFlags() (explainFlags, []string) {
	flags := explainFlags{
		depth: 1,
	}
	remainingArgs := []string{}
	args := os.Args[2:]

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--depth" || arg == "-d" {
			if i+1 < len(args) {
				if depth, err := strconv.Atoi(args[i+1]); err == nil {
					flags.depth = depth
				}
				i++
			}
		} else if arg == "--verbose" || arg == "-v" {
			flags.verbose = true
		} else if !strings.HasPrefix(arg, "-") {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, remainingArgs
}

func (a *App) handleExplain() {
	if a.config.CurrentConnection == "" {
		printError(
			"No active connection. Use 'squix switch <connection>' or 'squix init' first",
		)
	}

	flags, args := parseExplainFlags()

	if len(args) == 0 {
		fmt.Println("Usage: squix explain [--depth|-d N] [--verbose|-v] <table-name>")
		os.Exit(1)
	}

	conn := a.config.LiveConnection(a.config.CurrentConnection)

	if err := conn.Open(); err != nil {
		printError(
			"Could not open connection to %s: %v",
			a.config.CurrentConnection,
			err,
		)
	}
	defer conn.Close()

	tableName := args[0]

	fmt.Println(explain.BuildTree(conn, tableName, flags.depth, flags.verbose))
}
