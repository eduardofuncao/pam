package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/eduardofuncao/squix/internal/db"
)

func TestGroupKey(t *testing.T) {
	cases := []struct{ name, want string }{
		{"ecommerce:dev", "ecommerce"},
		{"ecommerce:prod", "ecommerce"},
		{"analytics", "analytics"},
		{"a/b:c", "a_b"}, // QueryGroupKey -> "a/b", then sanitized -> "a_b"
		{"weird name", "weird_name"},
	}
	for _, c := range cases {
		if got := GroupKey(c.name); got != c.want {
			t.Errorf("GroupKey(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseGroupFile_Basic(t *testing.T) {
	content := []byte("-- @query a\nSELECT 1\n\n-- @query b\nSELECT 2\n")
	got, err := ParseGroupFile(content)
	if err != nil {
		t.Fatalf("ParseGroupFile: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 queries, got %d: %+v", len(got), got)
	}
	if got["a"].SQL != "SELECT 1" {
		t.Errorf("a.SQL = %q", got["a"].SQL)
	}
	if got["b"].SQL != "SELECT 2" {
		t.Errorf("b.SQL = %q", got["b"].SQL)
	}
}

func TestParseGroupFile_TablePk(t *testing.T) {
	content := []byte("-- @query users\n-- @table users\n-- @pk id,org_id\nSELECT * FROM users\n")
	got, err := ParseGroupFile(content)
	if err != nil {
		t.Fatalf("ParseGroupFile: %v", err)
	}
	q := got["users"]
	if q.TableName != "users" {
		t.Errorf("TableName = %q", q.TableName)
	}
	if !reflect.DeepEqual(q.PrimaryKeys, []string{"id", "org_id"}) {
		t.Errorf("PrimaryKeys = %v", q.PrimaryKeys)
	}
}

func TestParseGroupFile_PreambleIgnored(t *testing.T) {
	// Header/help comments before the first block are ignored, even with no
	// blank line between preamble and the first -- @query.
	content := []byte("-- Editing queries for: foo\n-- @query a\nSELECT 1\n")
	got, err := ParseGroupFile(content)
	if err != nil {
		t.Fatalf("ParseGroupFile: %v", err)
	}
	if len(got) != 1 || got["a"].SQL != "SELECT 1" {
		t.Fatalf("expected single query a, got %+v", got)
	}
}

func TestParseGroupFile_MidSQLDirectiveStaysInBody(t *testing.T) {
	// A "-- @query" inside a body without a preceding blank line is SQL, not a
	// new block.
	content := []byte("-- @query a\nSELECT 1\n-- @query b\nSELECT 2\n")
	got, err := ParseGroupFile(content)
	if err != nil {
		t.Fatalf("ParseGroupFile: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 query (b absorbed into a), got %d: %+v", len(got), got)
	}
	if !contains(got["a"].SQL, "-- @query b") {
		t.Errorf("expected directive kept in body, got %q", got["a"].SQL)
	}
}

func TestParseGroupFile_MissingName(t *testing.T) {
	if _, err := ParseGroupFile([]byte("-- @query\nSELECT 1\n")); err == nil {
		t.Error("expected error for missing query name")
	}
}

func TestAssignIDs(t *testing.T) {
	queries := map[string]db.Query{
		"zeta":  {},
		"alpha": {},
		"mid":   {},
	}
	AssignIDs(queries)
	if queries["alpha"].Id != 1 || queries["mid"].Id != 2 || queries["zeta"].Id != 3 {
		t.Errorf("ids not assigned in sorted order: %+v", queries)
	}
}

func TestRenderGroupSorted(t *testing.T) {
	queries := map[string]db.Query{
		"b": {Name: "b", SQL: "SELECT 2"},
		"a": {Name: "a", SQL: "SELECT 1", TableName: "t", PrimaryKeys: []string{"id"}},
	}
	out := RenderGroup(queries)
	// a before b (sorted).
	if idxA, idxB := index(out, "-- @query a"), index(out, "-- @query b"); idxA < 0 || idxB < 0 || idxA > idxB {
		t.Errorf("expected a before b:\n%s", out)
	}
	if !contains(out, "-- @table t") || !contains(out, "-- @pk id") {
		t.Errorf("expected table/pk directives:\n%s", out)
	}
}

func TestRoundTrip(t *testing.T) {
	original := map[string]db.Query{
		"list_orders": {Name: "list_orders", SQL: "SELECT * FROM orders ORDER BY created_at"},
		"users":       {Name: "users", SQL: "SELECT id FROM users", TableName: "users", PrimaryKeys: []string{"id"}},
	}
	rendered := RenderGroup(original)
	parsed, err := ParseGroupFile([]byte(rendered))
	if err != nil {
		t.Fatalf("ParseGroupFile: %v", err)
	}
	if len(parsed) != len(original) {
		t.Fatalf("expected %d queries, got %d", len(original), len(parsed))
	}
	for name, want := range original {
		got, ok := parsed[name]
		if !ok {
			t.Errorf("query %q lost in round-trip", name)
			continue
		}
		if got.SQL != want.SQL || got.TableName != want.TableName ||
			!reflect.DeepEqual(got.PrimaryKeys, want.PrimaryKeys) {
			t.Errorf("query %q mismatch:\nwant %+v\ngot  %+v", name, want, got)
		}
	}
}

func TestWriteGroupFile(t *testing.T) {
	withTempCfg(t)
	queries := map[string]db.Query{"a": {Name: "a", SQL: "SELECT 1"}}
	if err := WriteGroupFile("grp", queries); err != nil {
		t.Fatalf("WriteGroupFile: %v", err)
	}
	data, err := os.ReadFile(GroupFile("grp"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !contains(string(data), "-- @query a") {
		t.Errorf("file missing rendered query:\n%s", data)
	}
}

func contains(s, sub string) bool { return index(s, sub) >= 0 }

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
