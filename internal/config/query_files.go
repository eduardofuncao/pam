package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eduardofuncao/squix/internal/db"
)

// QueriesDir is the directory holding one .sql file per query group.
func QueriesDir() string {
	return filepath.Join(CfgPath, "queries")
}

// GroupFile returns the path to the .sql file for a group key (already
// sanitized, no extension).
func GroupFile(key string) string {
	return filepath.Join(QueriesDir(), key+".sql")
}

// sanitizeFilename replaces filesystem-unsafe runes with '_'.
func sanitizeFilename(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9', r == '.', r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "_"
	}
	return b.String()
}

// GroupKey derives the shared query-library key for a connection name (via
// QueryGroupKey) and makes it filesystem-safe. This is both the in-memory
// QueryGroups map key and (with a .sql suffix) the filename, so every place a
// group meets memory or disk stays consistent.
func GroupKey(connName string) string {
	return sanitizeFilename(QueryGroupKey(connName))
}

// ParseGroupFile parses directive-delimited SQL into queries (without ids).
//
// Format:
//
//	-- @query <name>
//	[-- @table <table>]
//	[-- @pk <col>[,<col>...]]
//	<sql body until next block or EOF>
//
// Content before the first "-- @query" is ignored (preamble). A "-- @query"
// line starts a new block only at BOF or when preceded by a blank line, so a
// stray "-- @query" mid-SQL (no blank line before it) stays part of the body.
func ParseGroupFile(content []byte) (map[string]db.Query, error) {
	queries := make(map[string]db.Query)

	var (
		name      string
		table     string
		pk        []string
		body      strings.Builder
		inBlock   bool
		inHeader  bool
		prevBlank = true // BOF counts as blank-preceded
	)

	flush := func() {
		if !inBlock {
			return
		}
		sql := strings.TrimSpace(body.String())
		if name == "" {
			return
		}
		q := db.Query{Name: name, SQL: sql}
		if table != "" {
			q.TableName = table
		}
		if len(pk) > 0 {
			q.PrimaryKeys = pk
		}
		queries[name] = q
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		// Block start: "-- @query <name>". Before the first block it always
		// starts one (no body to protect); once inside a block it must be
		// blank-preceded so a stray "-- @query" mid-SQL stays in the body.
		if rest, ok := directive(trimmed, "@query"); ok && (prevBlank || !inBlock) {
			flush()
			name = strings.TrimSpace(rest)
			if name == "" {
				return nil, fmt.Errorf("missing query name after '-- @query'")
			}
			table = ""
			pk = nil
			body.Reset()
			inBlock = true
			inHeader = true
			prevBlank = false
			continue
		}

		if inBlock && inHeader {
			if rest, ok := directive(trimmed, "@table"); ok {
				table = strings.TrimSpace(rest)
				prevBlank = false
				continue
			}
			if rest, ok := directive(trimmed, "@pk"); ok {
				for c := range strings.SplitSeq(rest, ",") {
					if c = strings.TrimSpace(c); c != "" {
						pk = append(pk, c)
					}
				}
				prevBlank = false
				continue
			}
			// Blank lines inside the header are tolerated; first real line
			// begins the SQL body.
			if trimmed == "" {
				prevBlank = true
				continue
			}
			inHeader = false
		}

		if inBlock && !inHeader {
			if body.Len() > 0 {
				body.WriteByte('\n')
			}
			body.WriteString(line)
		}

		prevBlank = trimmed == ""
	}
	flush()

	return queries, nil
}

// directive reports whether trimmed is a "-- @<name>" line and returns the
// trimmed text following the directive.
func directive(trimmed, name string) (string, bool) {
	prefix := "-- " + name
	rest, ok := strings.CutPrefix(trimmed, prefix)
	if !ok {
		return "", false
	}
	// Require a space or end-of-line after the directive name.
	if rest == "" {
		return "", true
	}
	if rest[0] == ' ' || rest[0] == '\t' {
		return rest, true
	}
	return "", false
}

// AssignIDs assigns sequential ids (1-based) to queries sorted by name.
func AssignIDs(queries map[string]db.Query) {
	names := make([]string, 0, len(queries))
	for n := range queries {
		names = append(names, n)
	}
	sort.Strings(names)
	for i, n := range names {
		q := queries[n]
		q.Id = i + 1
		queries[n] = q
	}
}

// RenderQuery renders a single query block (no leading/trailing blank lines).
func RenderQuery(q db.Query) string {
	var b strings.Builder
	fmt.Fprintf(&b, "-- @query %s\n", q.Name)
	if q.TableName != "" {
		fmt.Fprintf(&b, "-- @table %s\n", q.TableName)
	}
	if len(q.PrimaryKeys) > 0 {
		fmt.Fprintf(&b, "-- @pk %s\n", strings.Join(q.PrimaryKeys, ","))
	}
	b.WriteString(strings.TrimSpace(q.SQL))
	return b.String()
}

// RenderGroup renders all queries in a group, sorted by name, one block per
// query separated by blank lines.
func RenderGroup(queries map[string]db.Query) string {
	names := make([]string, 0, len(queries))
	for n := range queries {
		names = append(names, n)
	}
	sort.Strings(names)

	blocks := make([]string, 0, len(names))
	for _, n := range names {
		blocks = append(blocks, RenderQuery(queries[n]))
	}
	return strings.Join(blocks, "\n\n") + "\n"
}

// WriteGroupFile atomically writes a group's queries to its .sql file.
func WriteGroupFile(key string, queries map[string]db.Query) error {
	if err := os.MkdirAll(QueriesDir(), 0755); err != nil {
		return err
	}
	path := GroupFile(key)
	tmp, err := os.CreateTemp(QueriesDir(), ".tmp-*.sql")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.WriteString(RenderGroup(queries)); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
