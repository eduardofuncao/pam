package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eduardofuncao/squix/internal/backup"
	"github.com/eduardofuncao/squix/internal/styles"
)

func (a *App) handleExport() {
	if len(os.Args) < 3 {
		printError("usage: squix export [backup|schema]")
	}

	switch os.Args[2] {
	case "backup":
		a.runExport(os.Args[3:], false, "Backup")
	case "schema":
		a.runExport(os.Args[3:], true, "Schema")
	default:
		printError("usage: squix export [backup|schema]")
	}
}

func (a *App) runExport(args []string, schemaOnly bool, label string) {
	var path string
	var format string
	var tables []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format":
			if i+1 < len(args) {
				i++
				format = args[i]
			}
		case "--table":
			if i+1 < len(args) {
				i++
				tables = append(tables, args[i])
			}
		default:
			if path == "" && arg[0] != '-' {
				path = arg
			}
		}
	}

	if !schemaOnly && len(tables) > 0 {
		printError("--table is only valid for export schema")
	}

	if a.config.CurrentConnection == "" {
		printError("No active connection. Use 'squix switch <connection>' or 'squix init' first")
	}

	conn := a.config.Connections[a.config.CurrentConnection]
	dbType := conn.DBType
	connString := os.ExpandEnv(conn.ConnString)

	dumper, err := backup.CreateDumper(dbType)
	if err != nil {
		printError("%v", err)
	}

	fileExt := filepath.Ext(path)
	if schemaOnly && fileExt == "" && format == "" {
		format = "plain"
	}

	formatName, err := backup.ResolveFormat(fileExt, format, dumper)
	if err != nil {
		printError("%v", err)
	}

	spec := dumper.Formats()[formatName]
	outPath := backup.ResolvePath(path, a.config.CurrentConnection, spec.Ext)

	opts := backup.DumpOptions{
		Format:     formatName,
		OutPath:    outPath,
		SchemaOnly: schemaOnly,
		Tables:     tables,
	}

	if err := dumper.Dump(connString, opts); err != nil {
		printError("%v", err)
	}

	fmt.Println(styles.Success.Render(fmt.Sprintf("✓ %s written to %s", label, outPath)))
}
