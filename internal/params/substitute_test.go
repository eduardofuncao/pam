package params

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/eduardofuncao/squix/internal/db"
)

type mockConn struct {
	db.BaseConnection
	placeholder string
}

func (m *mockConn) GetPlaceholder(idx int) string                 { return m.placeholder }
func (m *mockConn) GetUniqueConstraints(string) ([]string, error) { return nil, nil }

func newMockConn(placeholder string) db.DatabaseConnection {
	return &mockConn{
		BaseConnection: db.BaseConnection{},
		placeholder:    placeholder,
	}
}

// needed because BaseConnection.Query returns wrong error message
func (m *mockConn) Query(string, ...any) (any, error)                  { return nil, nil }
func (m *mockConn) ExecQuery(string, ...any) (*sql.Rows, error)        { return nil, nil }
func (m *mockConn) Exec(string, ...any) (sql.Result, error)            { return nil, nil }
func (m *mockConn) Open() error                                        { return nil }
func (m *mockConn) Ping() error                                        { return nil }
func (m *mockConn) Close() error                                       { return nil }
func (m *mockConn) GetTableMetadata(string) (*db.TableMetadata, error) { return nil, nil }
func (m *mockConn) GetTables() ([]string, error)                       { return nil, nil }
func (m *mockConn) GetViews() ([]string, error)                        { return nil, nil }
func (m *mockConn) GetForeignKeys(string) ([]db.ForeignKey, error)     { return nil, nil }
func (m *mockConn) GetForeignKeysReferencingTable(string) ([]db.ForeignKey, error) {
	return nil, nil
}

func TestSubstituteParameters(t *testing.T) {
	t.Run("no params", func(t *testing.T) {
		sql, args, err := SubstituteParameters("SELECT 1", nil, newMockConn("?"))
		if err != nil {
			t.Fatal(err)
		}
		if sql != "SELECT 1" {
			t.Errorf("got %q", sql)
		}
		if len(args) != 0 {
			t.Errorf("expected 0 args, got %d", len(args))
		}
	})

	t.Run("single param", func(t *testing.T) {
		sql, args, err := SubstituteParameters(
			"SELECT * FROM t WHERE id = :id",
			map[string]string{"id": "1"},
			newMockConn("?"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(sql, "?") {
			t.Errorf("expected placeholder in sql: %q", sql)
		}
		if len(args) != 1 || args[0] != "1" {
			t.Errorf("expected args [1], got %v", args)
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		sql, args, err := SubstituteParameters(
			"SELECT * FROM t WHERE name = :name AND age > :age",
			map[string]string{"name": "Alice", "age": "25"},
			newMockConn("?"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(sql, "?") != 2 {
			t.Errorf("expected 2 placeholders in sql: %q", sql)
		}
		if len(args) != 2 {
			t.Errorf("expected 2 args, got %d", len(args))
		}
	})

	t.Run("repeated param", func(t *testing.T) {
		sql, _, err := SubstituteParameters(
			"SELECT * FROM t WHERE x = :id AND y = :id",
			map[string]string{"id": "1"},
			newMockConn("?"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(sql, "?") != 2 {
			t.Errorf("expected 2 placeholders: %q", sql)
		}
	})

	t.Run("missing param value", func(t *testing.T) {
		_, _, err := SubstituteParameters(
			"SELECT * FROM t WHERE id = :id AND name = :name",
			map[string]string{"name": "Alice"},
			newMockConn("?"),
		)
		if err == nil {
			t.Error("expected error for missing param")
		}
		if !strings.Contains(err.Error(), "missing value") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("postgres-style placeholder", func(t *testing.T) {
		sql, _, err := SubstituteParameters(
			"SELECT * FROM t WHERE id = :id",
			map[string]string{"id": "1"},
			newMockConn("$1"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(sql, "$1") {
			t.Errorf("expected $1 placeholder: %q", sql)
		}
	})

	t.Run("no matches in SQL", func(t *testing.T) {
		sql, args, err := SubstituteParameters(
			"SELECT 1",
			map[string]string{"id": "1"},
			newMockConn("?"),
		)
		if err != nil {
			t.Fatal(err)
		}
		if sql != "SELECT 1" {
			t.Errorf("got %q", sql)
		}
		if len(args) != 0 {
			t.Errorf("expected 0 args")
		}
	})
}

func TestGenerateDisplaySQL(t *testing.T) {
	t.Run("numeric value quoted", func(t *testing.T) {
		got := GenerateDisplaySQL(
			"SELECT * FROM t WHERE id = :id",
			map[string]string{"id": "42"},
		)
		if !strings.Contains(got, "'42'") {
			t.Errorf("expected quoted '42' in output: %q", got)
		}
	})

	t.Run("string value quoted and escaped", func(t *testing.T) {
		got := GenerateDisplaySQL(
			"SELECT * FROM t WHERE name = :name",
			map[string]string{"name": "O'Brien"},
		)
		if !strings.Contains(got, "'O''Brien'") {
			t.Errorf("expected escaped quoted value: %q", got)
		}
	})

	t.Run("no matching param", func(t *testing.T) {
		sql := "SELECT * FROM t WHERE id = :id"
		got := GenerateDisplaySQL(sql, map[string]string{})
		if got != sql {
			t.Errorf("expected unchanged sql, got %q", got)
		}
	})
}
