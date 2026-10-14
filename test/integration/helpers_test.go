package integration

import (
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

var binaryPath string

func TestMain(m *testing.M) {
	wd, _ := os.Getwd()
	projectRoot := filepath.Clean(filepath.Join(wd, "..", ".."))

	tmp, err := os.MkdirTemp("", "squix-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	binaryPath = filepath.Join(tmp, "squix")

	cmd := exec.Command("go", "build", "-ldflags=-s -w", "-o", binaryPath, "./cmd/squix")
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build squix (dir=%s): %v\n%s\n", projectRoot, err, out)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

type TestEnv struct {
	HomeDir string
	DBPath  string
	t       *testing.T
}

func Setup(t *testing.T) *TestEnv {
	t.Helper()
	home := t.TempDir()
	configDir := filepath.Join(home, ".config", "squix")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")

	config := fmt.Sprintf(
		"current_connection: testdb\nconnections:\n  testdb:\n    name: testdb\n    db_type: sqlite\n    conn_string: %s\n",
		dbPath,
	)
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return &TestEnv{HomeDir: home, DBPath: dbPath, t: t}
}

const DefaultSchema = `
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    age INTEGER
);
CREATE TABLE orders (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product TEXT NOT NULL,
    amount REAL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`

var DefaultInserts = []string{
	"INSERT INTO users (name, email, age) VALUES ('Alice', 'alice@example.com', 30)",
	"INSERT INTO users (name, email, age) VALUES ('Bob', 'bob@example.com', 25)",
	"INSERT INTO users (name, email, age) VALUES ('Charlie', 'charlie@example.com', 35)",
	"INSERT INTO orders (user_id, product, amount) VALUES (1, 'Widget', 9.99)",
	"INSERT INTO orders (user_id, product, amount) VALUES (2, 'Gadget', 24.50)",
	"INSERT INTO orders (user_id, product, amount) VALUES (1, 'Doohickey', 5.00)",
}

func (e *TestEnv) SeedData(schema string, inserts []string) {
	e.t.Helper()
	db, err := sql.Open("sqlite", e.DBPath)
	if err != nil {
		e.t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(schema); err != nil {
		e.t.Fatalf("exec schema: %v", err)
	}
	for _, ins := range inserts {
		if _, err := db.Exec(ins); err != nil {
			e.t.Fatalf("exec insert: %v\n%s", err, ins)
		}
	}
}

func (e *TestEnv) SeedDefaults() {
	e.t.Helper()
	e.SeedData(DefaultSchema, DefaultInserts)
}

// WriteConfig overwrites config.yaml in the env's config dir with the given YAML.
func (e *TestEnv) WriteConfig(yaml string) {
	e.t.Helper()
	path := filepath.Join(e.HomeDir, ".config", "squix", "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		e.t.Fatalf("write config: %v", err)
	}
}

// SeedDB creates a sqlite database at path and seeds it with the default schema
// and inserts. Used to give multiple {service}:{env} connections distinct files.
func (e *TestEnv) SeedDB(path string) {
	e.t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		e.t.Fatalf("open sqlite: %v", err)
	}
	defer db.Close()
	if _, err := db.Exec(DefaultSchema); err != nil {
		e.t.Fatalf("exec schema: %v", err)
	}
	for _, ins := range DefaultInserts {
		if _, err := db.Exec(ins); err != nil {
			e.t.Fatalf("exec insert: %v\n%s", err, ins)
		}
	}
}

// NewDBPath returns a unique sqlite file path under the test's temp dir.
func (e *TestEnv) NewDBPath(name string) string {
	return filepath.Join(e.t.TempDir(), name)
}

func (e *TestEnv) RunSquix(args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+e.HomeDir)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			e.t.Logf("RunSquix(%v) exec error: %v (binary=%s)", args, err, binaryPath)
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}

func (e *TestEnv) RunSquixWithStdin(stdin string, args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+e.HomeDir)
	cmd.Stdin = strings.NewReader(stdin)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			e.t.Logf("RunSquixWithStdin(%v) exec error: %v", args, err)
			exitCode = -1
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}
