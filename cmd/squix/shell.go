package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/eduardofuncao/squix/internal/db"
	"github.com/eduardofuncao/squix/internal/params"
	"github.com/eduardofuncao/squix/internal/run"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) handleShell() {
	if a.config.CurrentConnection == "" {
		printError("No active connection.   Use 'squix switch <connection>' or 'squix init' first")
	}

	conn := a.config.LiveConnection(a.config.CurrentConnection)
	var err error
	if err = conn.Open(); err != nil {
		printError("Could not open connection: %v", err)
	}
	defer conn.Close()

	// Startup banner
	a.printConnStatus(conn)
	fmt.Println(styles.Faint.Render("Type queries (end with ;) or 'quit' to exit."))

	normalPrompt := styles.Title.Render(fmt.Sprintf("squix@%s> ", a.config.CurrentConnection))
	contPrompt := styles.Faint.Render("... ")

	rl, err := readline.NewEx(&readline.Config{
		Prompt: normalPrompt,
	})
	if err != nil {
		printError("readline error: %v", err)
	}
	defer rl.Close()

	var multiLine strings.Builder

	for {
		line, err := rl.Readline()
		if err == io.EOF || err == readline.ErrInterrupt {
			if multiLine.Len() > 0 {
				multiLine.Reset()
				rl.SetPrompt(normalPrompt)
				continue
			}
			break
		}
		if err != nil {
			break
		}

		trimmed := strings.TrimSpace(line)

		// Multi-line continuation
		if multiLine.Len() > 0 {
			if trimmed == "" {
				input := strings.TrimSpace(multiLine.String())
				multiLine.Reset()
				rl.SetPrompt(normalPrompt)
				if shouldExit := a.executeReplLine(rl, conn, input); shouldExit {
					return
				}
			} else {
				multiLine.WriteString(trimmed)
				multiLine.WriteString("\n")
				if strings.HasSuffix(trimmed, ";") {
					full := strings.TrimSpace(multiLine.String())
					full = strings.TrimSuffix(full, ";")
					multiLine.Reset()
					rl.SetPrompt(normalPrompt)
					if shouldExit := a.executeReplLine(rl, conn, full); shouldExit {
						return
					}
				} else {
					rl.SetPrompt(contPrompt)
				}
			}
			continue
		}

		if trimmed == "" {
			continue
		}

		// Start multi-line if SQL without trailing ;
		if run.IsLikelySQL(trimmed) && !strings.HasSuffix(trimmed, ";") {
			multiLine.WriteString(trimmed)
			multiLine.WriteString("\n")
			rl.SetPrompt(contPrompt)
			continue
		}

		trimmed = strings.TrimSuffix(trimmed, ";")

		if shouldExit := a.executeReplLine(rl, conn, trimmed); shouldExit {
			return
		}
	}

	fmt.Println(styles.Faint.Render("Done"))
}

// executeReplLine processes a single completed REPL line. Returns true if shell should exit.
func (a *App) executeReplLine(rl *readline.Instance, conn db.DatabaseConnection, input string) bool {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return false
	}

	switch strings.ToLower(trimmed) {
	case "exit", "quit", "\\q":
		fmt.Println(styles.Faint.Render("Done"))
		return true
	case "help", "\\h":
		fmt.Println(shellHelpText())
		return false
	case "status":
		a.printConnStatus(conn)
		return false
	}

	// Handle list commands with args
	tokens := parseInput(trimmed)
	if len(tokens) > 0 {
		switch tokens[0] {
		case "list", "ls", "\\l":
			a.handleListWithArgs(tokens[1:])
			return false
		case "explore", "tables", "\\dt":
			a.handleReplExplore(conn, tokens[1:])
			return false
		}
	}

	rl.SaveHistory(trimmed)

	if err := a.runFromArgsOpenConn(tokens, conn); err != nil {
		fmt.Println(styles.Error.Render(fmt.Sprintf("Error: %v", err)))
	}
	return false
}

// parseInput tokenizes shell-like input, respecting quoted strings.
// If the input looks like inline SQL, it returns the whole string as one arg.
func parseInput(input string) []string {
	trimmed := strings.TrimSpace(input)

	// If it looks like SQL, pass the whole thing as the selector
	if run.IsLikelySQL(trimmed) {
		return []string{trimmed}
	}

	// Otherwise, shell-like tokenization
	var tokens []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]
		switch {
		case ch == '"' && !inQuote:
			inQuote = true
		case ch == '"' && inQuote:
			inQuote = false
		case ch == ' ' && !inQuote:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// runFromArgsOpenConn is like runFromArgs but uses ExecuteWithOpenConn (no Open/Close)
func (a *App) runFromArgsOpenConn(args []string, conn db.DatabaseConnection) error {
	flags := parseRunFlagsFrom(args)

	resolved, err := run.ResolveQuery(flags, a.config.CurrentConnection, conn)
	if err != nil {
		return err
	}

	if run.ShouldCreateNewQuery(resolved) {
		return fmt.Errorf("no query specified. Use a query name, inline SQL, or --last")
	}

	if flags.EditMode && !flags.LastQuery {
		resolved.Query = a.editQueryOrExit(resolved.Query)
	}

	a.saveIfNeeded(resolved)

	paramFlags := parseParameterFlagsFrom(args)
	positionalArgsSlice := parsePositionalArgsFrom(args, flags.Selector)
	positionalArgs := params.MapPositionalArgs(resolved.Query.SQL, positionalArgsSlice)

	// If --format is set, use export executor
	if flags.ExportFormat != "" {
		return a.executeQueryWithParamsInternal(resolved.Query, conn, paramFlags, positionalArgs, func(p run.ExecutionParams) error {
			return run.ExecuteExportWithOpenConn(p, flags.ExportFormat)
		}, true, flags)
	}

	return a.executeQueryWithParamsInternal(resolved.Query, conn, paramFlags, positionalArgs, run.ExecuteWithOpenConn, false, flags)
}

func shellHelpText() string {
	var sb strings.Builder
	sb.WriteString(styles.Title.Render("Squix Shell Help"))
	sb.WriteString("\n\n")
	sb.WriteString(styles.Title.Render("Commands"))
	sb.WriteString("\n")
	sb.WriteString("  exit, quit, \\q       Exit the REPL\n")
	sb.WriteString("  help, \\h             Show this help\n")
	sb.WriteString("  status               Show connection info\n")
	sb.WriteString("  list, ls, \\l         List saved queries or connections\n")
	sb.WriteString("  explore, \\dt [name]  Open schema explorer or view table data\n")
	sb.WriteString("\n")
	sb.WriteString(styles.Title.Render("Query Execution"))
	sb.WriteString("\n")
	sb.WriteString("  <query-name>           Run a saved query by name\n")
	sb.WriteString("  <query-name> val1      Run with positional params\n")
	sb.WriteString("  <query-name> --p val   Run with named params\n")
	sb.WriteString("  select ...             Run inline SQL\n")
	sb.WriteString("  --last, -l             Rerun last query\n")
	sb.WriteString("  -e <query>             Edit query before running\n")
	sb.WriteString("\n")
	sb.WriteString(styles.Title.Render("Multi-line"))
	sb.WriteString("\n")
	sb.WriteString("  Type SQL without trailing ; to enter multi-line mode.\n")
	sb.WriteString("  Add ; at end of the line or press enter on an empty line to execute.\n")
	sb.WriteString("  Ctrl+C to cancel query.\n")
	return sb.String()
}
