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
	depth int
}

func parseExplainFlags() (explainFlags, []string) {
	flags := explainFlags{
		depth: 1, // Default to showing just direct relationships
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
				i++ // Skip the next argument
			}
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
		fmt.Println("Usage: pam explain [--depth|-d N] <table-name>")
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
	seenTables := make(map[string]bool)

	// Get "belongs to" relationships (FKs from this table to other tables)
	belongsToFKs, err := conn.GetForeignKeys(tableName)
	if err == nil {
		for _, fk := range belongsToFKs {
			// Create a unique key for deduplication
			key := fmt.Sprintf("%s:%s:%s", fk.ReferencedTable, fk.Column, fk.ReferencedColumn)
			if seenTables[key] {
				continue
			}
			seenTables[key] = true
			relationships = append(relationships, relationship{
				relType:           belongsTo,
				column:            fk.Column,
				referencedTable:   fk.ReferencedTable,
				referencedColumn:  fk.ReferencedColumn,
			})
		}
	}

	// Get "has many" relationships (FKs from other tables to this table)
	hasManyFKs, err := conn.GetForeignKeysReferencingTable(tableName)
	if err == nil {
		for _, fk := range hasManyFKs {
			// Create a unique key for deduplication
			key := fmt.Sprintf("%s:%s:%s", fk.ReferencedTable, fk.Column, fk.ReferencedColumn)
			if seenTables[key] {
				continue
			}
			seenTables[key] = true
			relationships = append(relationships, relationship{
				relType:           hasMany,
				column:            fk.Column,
				referencedTable:   fk.ReferencedTable,
				referencedColumn:  fk.ReferencedColumn,
			})
		}
	}

	return a.renderNode(conn, tableName, relationships, maxDepth, currentDepth, visited, fkCache, true)
}

func (a *App) renderNode(conn db.DatabaseConnection, tableName string, relationships []relationship, maxDepth, currentDepth int, visited map[string]bool, fkCache map[string][]db.ForeignKey, isRoot bool) string {
	var builder strings.Builder

	// Render current table with PK info (only at root level)
	if isRoot {
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

	// Mark current table as visited to prevent cycles
	visited[tableName] = true

	// Check if we should render relationships at this depth
	if currentDepth >= maxDepth && !isRoot {
		return builder.String()
	}

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

		// Don't show children for self-references
		if isSelfReference {
			continue
		}

		// Don't revisit tables
		if visited[rel.referencedTable] {
			continue
		}

		// Recursively render children if within depth limit
		if currentDepth+1 <= maxDepth {
			childRelationships := a.getChildRelationships(conn, rel.referencedTable)
			if len(childRelationships) > 0 {
				childPrefix := "    "
				if !isLast {
					childPrefix = "│   "
				}

				// Create local visited map for this branch to prevent cycles
				localVisited := make(map[string]bool)
				for k, v := range visited {
					localVisited[k] = v
				}
				localVisited[rel.referencedTable] = true

				childTree := a.renderNode(conn, rel.referencedTable, childRelationships, maxDepth, currentDepth+1, localVisited, fkCache, false)
				lines := strings.Split(childTree, "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					builder.WriteString(styles.TreeConnector.Render(childPrefix))
					builder.WriteString(line)
					builder.WriteString("\n")
				}
			}
		}
	}

	return builder.String()
}

func (a *App) getChildRelationships(conn db.DatabaseConnection, tableName string) []relationship {
	var relationships []relationship
	seenTables := make(map[string]bool)

	// Get "belongs to" relationships
	belongsToFKs, err := conn.GetForeignKeys(tableName)
	if err == nil {
		for _, fk := range belongsToFKs {
			key := fmt.Sprintf("%s:%s:%s", fk.ReferencedTable, fk.Column, fk.ReferencedColumn)
			if !seenTables[key] {
				seenTables[key] = true
				relationships = append(relationships, relationship{
					relType:           belongsTo,
					column:            fk.Column,
					referencedTable:   fk.ReferencedTable,
					referencedColumn:  fk.ReferencedColumn,
				})
			}
		}
	}

	// Get "has many" relationships
	hasManyFKs, err := conn.GetForeignKeysReferencingTable(tableName)
	if err == nil {
		for _, fk := range hasManyFKs {
			key := fmt.Sprintf("%s:%s:%s", fk.ReferencedTable, fk.Column, fk.ReferencedColumn)
			if !seenTables[key] {
				seenTables[key] = true
				relationships = append(relationships, relationship{
					relType:           hasMany,
					column:            fk.Column,
					referencedTable:   fk.ReferencedTable,
					referencedColumn:  fk.ReferencedColumn,
				})
			}
		}
	}

	return relationships
}

func (a *App) renderRelationshipLine(rel relationship, parentTable string, isSelfReference bool) string {
	var builder strings.Builder

	var relText, cardinality, fkDetails string
	var relStyle lipgloss.Style

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

	return builder.String()
}
