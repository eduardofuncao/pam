package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/eduardofuncao/pam/internal/config"
	"github.com/eduardofuncao/pam/internal/db"
	"github.com/eduardofuncao/pam/internal/styles"
)

type explainFlags struct {
	depth       int
	showColumns bool
}

func parseExplainFlags() (explainFlags, []string) {
	flags := explainFlags{
		depth:       3,
		showColumns: false,
	}
	remainingArgs := []string{}
	args := os.Args[2:]

	for i, arg := range args {
		if arg == "--depth" || arg == "-d" {
			if i+1 < len(args) {
				if depth, err := strconv.Atoi(args[i+1]); err == nil {
					flags.depth = depth
				}
			}
		} else if arg == "--columns" || arg == "-c" {
			flags.showColumns = true
		} else if !strings.HasPrefix(arg, "-") {
			remainingArgs = append(remainingArgs, arg)
		}
	}

	return flags, remainingArgs
}

func (a *App) handleExplain() {
	if a.config.CurrentConnection == "" {
		printError(
			"No active connection. Use 'pam switch <connection>' or 'pam init' first",
		)
	}

	flags, args := parseExplainFlags()

	if len(args) == 0 {
		fmt.Println("Usage: pam explain [--depth|-d N] [--columns|-c] <table-name>")
		os.Exit(1)
	}

	conn := config.FromConnectionYaml(
		a.config.Connections[a.config.CurrentConnection],
	)

	if err := conn.Open(); err != nil {
		printError(
			"Could not open connection to %s: %v",
			a.config.CurrentConnection,
			err,
		)
	}
	defer conn.Close()

	tableName := args[0]

	// Cache for FK lookups to improve performance
	fkCache := make(map[string][]db.ForeignKey)
	visited := make(map[string]bool)

	tree := a.buildRelationshipTree(conn, tableName, flags.depth, 0, visited, fkCache)

	fmt.Println(tree)
}

type relationshipType int

const (
	belongsTo relationshipType = iota
	hasMany
)

type relationship struct {
	relType           relationshipType
	column            string
	referencedTable   string
	referencedColumn  string
}

func (a *App) buildRelationshipTree(conn db.DatabaseConnection, tableName string, maxDepth, currentDepth int, visited map[string]bool, fkCache map[string][]db.ForeignKey) string {
	if currentDepth >= maxDepth {
		return ""
	}

	var relationships []relationship

	// Only show "belongs to" relationships (FKs from this table to other tables)
	belongsToFKs, err := conn.GetForeignKeys(tableName)
	if err == nil {
		for _, fk := range belongsToFKs {
			relationships = append(relationships, relationship{
				relType:           belongsTo,
				column:            fk.Column,
				referencedTable:   fk.ReferencedTable,
				referencedColumn:  fk.ReferencedColumn,
			})
		}
	}

	return a.renderNode(conn, tableName, relationships, maxDepth, currentDepth, visited, fkCache)
}

func (a *App) renderNode(conn db.DatabaseConnection, tableName string, relationships []relationship, maxDepth, currentDepth int, visited map[string]bool, fkCache map[string][]db.ForeignKey) string {
	var builder strings.Builder

	// Render current table with PK info
	if currentDepth == 0 {
		metadata, _ := conn.GetTableMetadata(tableName)
		if len(metadata.PrimaryKeys) > 0 {
			pks := strings.Join(metadata.PrimaryKeys, ", ")
			builder.WriteString(styles.TableName.Render(tableName))
			builder.WriteString(" ")
			builder.WriteString(styles.PrimaryKeyLabel.Render(fmt.Sprintf("(PK: %s)", pks)))
		} else {
			builder.WriteString(styles.TableName.Render(tableName))
		}
		builder.WriteString("\n")
	}

	visited[tableName] = true

	for i, rel := range relationships {
		isLast := i == len(relationships)-1
		prefix := "├── "
		if isLast {
			prefix = "└── "
		}

		// Determine relationship type and styling
		var relText, cardinality, fkDetails string
		var relStyle lipgloss.Style

		isSelfReference := (rel.referencedTable == tableName)

		if rel.relType == belongsTo {
			relText = "belongs to"
			cardinality = "[N:1]"
			relStyle = styles.BelongsToStyle
			fkDetails = fmt.Sprintf("(FK: %s → %s.%s)", rel.column, rel.referencedTable, rel.referencedColumn)
		} else {
			relText = "has many"
			cardinality = "[1:N]"
			relStyle = styles.HasManyStyle
			fkDetails = fmt.Sprintf("(on: %s ← %s.%s)", rel.referencedColumn, rel.referencedTable, rel.column)
		}

		// Render relationship line
		builder.WriteString(styles.TreeConnector.Render(prefix))
		builder.WriteString(relStyle.Render(fmt.Sprintf("%s →", relText)))
		builder.WriteString(" ")
		builder.WriteString(styles.CardinalityStyle.Render(cardinality))
		builder.WriteString(" ")
		builder.WriteString(styles.TableName.Render(rel.referencedTable))
		builder.WriteString(" ")
		builder.WriteString(styles.Faint.Render(fkDetails))

		if isSelfReference {
			builder.WriteString(" " + styles.Faint.Render("(self-reference)"))
		}

		builder.WriteString("\n")

		// Don't recurse into self-references
		if isSelfReference {
			continue
		}

		// Don't revisit tables
		if visited[rel.referencedTable] {
			continue
		}

		// Build child relationships
		childPrefix := "    "
		if !isLast {
			childPrefix = "│   "
		}

		childTree := a.buildRelationshipTree(conn, rel.referencedTable, maxDepth, currentDepth+1, visited, fkCache)
		if childTree != "" {
			lines := strings.Split(childTree, "\n")
			for j, line := range lines {
				if line == "" {
					continue
				}
				if j > 0 {
					builder.WriteString(styles.TreeConnector.Render(childPrefix))
				}
				builder.WriteString(line + "\n")
			}
		}
	}

	return builder.String()
}
