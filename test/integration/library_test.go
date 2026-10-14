package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// multiConnConfig builds a config.yaml string with the given connections
// (name -> sqlite conn_string) and the given current connection. Query groups
// are intentionally omitted so tests exercise the live code path.
func multiConnConfig(conns map[string]string, current string) string {
	var b strings.Builder
	b.WriteString("current_connection: " + current + "\nconnections:\n")
	for name, cs := range conns {
		b.WriteString("  " + name + ":\n")
		b.WriteString("    name: " + name + "\n")
		b.WriteString("    db_type: sqlite\n")
		b.WriteString("    conn_string: " + cs + "\n")
	}
	return b.String()
}

func TestSharedLibraryAcrossEnvs(t *testing.T) {
	env := Setup(t)
	devPath := env.NewDBPath("dev.db")
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev":  devPath,
		"ecommerce:prod": env.NewDBPath("prod.db"),
	}, "ecommerce:dev"))

	if _, _, code := env.RunSquix("add", "list_users", "SELECT * FROM users"); code != 0 {
		t.Fatal("add failed")
	}

	// Switch to prod (sibling). The query added on dev must be visible.
	env.RunSquix("switch", "ecommerce:prod")
	stdout, _, code := env.RunSquix("list", "queries")
	if code != 0 {
		t.Fatalf("list queries failed: %s", stdout)
	}
	if !strings.Contains(stdout, "list_users") {
		t.Errorf("prod does not see query shared from dev:\n%s", stdout)
	}
}

func TestStandalonePrivateLibrary(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev": env.NewDBPath("dev.db"),
		"analytics":     env.NewDBPath("analytics.db"),
	}, "ecommerce:dev"))

	env.RunSquix("add", "list_users", "SELECT * FROM users")

	env.RunSquix("switch", "analytics")
	stdout, _, _ := env.RunSquix("list", "queries")
	if strings.Contains(stdout, "list_users") {
		t.Errorf("standalone connection leaked query from ecommerce group:\n%s", stdout)
	}
}

func TestLastQueryClearedOnSwitch(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev":  env.NewDBPath("dev.db"),
		"ecommerce:prod": env.NewDBPath("prod.db"),
	}, "ecommerce:dev"))

	// Run a query on dev → writes last-query.sql via the execute chokepoint.
	if _, _, code := env.RunSquix("run", "SELECT 1", "-f", "csv"); code != 0 {
		t.Fatal("run on dev failed")
	}

	// Switching connections clears last-query.sql (ephemeral, no longer stored
	// per-connection in config.yaml).
	env.RunSquix("switch", "ecommerce:prod")

	// --last on prod must fail: switching wiped the last query.
	_, stderr, code := env.RunSquix("run", "--last")
	if code == 0 {
		t.Fatalf("run --last after switch should fail; stderr: %s", stderr)
	}
	if !strings.Contains(strings.ToLower(stderr), "no last query") {
		t.Errorf("expected 'no last query' error, got stderr: %s", stderr)
	}
}

func TestListConnections_AnnotatesGroup(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev": env.NewDBPath("dev.db"),
		"analytics":     env.NewDBPath("analytics.db"),
	}, "ecommerce:dev"))

	stdout, _, code := env.RunSquix("list", "connections")
	if code != 0 {
		t.Fatalf("list connections failed: %s", stdout)
	}
	// The ecommerce:dev line shows its shared group key "ecommerce".
	devLine := ""
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "ecommerce:dev") {
			devLine = line
			break
		}
	}
	if devLine == "" {
		t.Fatalf("ecommerce:dev not listed:\n%s", stdout)
	}
	// The group key appears (after the db_type). Count "ecommerce" occurrences
	// on the line: name "ecommerce:dev" is one, the annotation is another.
	if strings.Count(devLine, "ecommerce") < 2 {
		t.Errorf("expected group annotation on dev line, got: %q", devLine)
	}
}

func TestInitSibling_SeesExistingGroup(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev": env.NewDBPath("dev.db"),
	}, "ecommerce:dev"))

	env.RunSquix("add", "list_users", "SELECT * FROM users")

	// init a sibling env; its sqlite file is created on connect.
	prodPath := env.NewDBPath("prod.db")
	if _, _, code := env.RunSquix("init", "ecommerce:prod", prodPath); code != 0 {
		t.Fatal("init sibling failed")
	}

	// init sets current to the new connection.
	stdout, _, code := env.RunSquix("list", "queries")
	if code != 0 {
		t.Fatalf("list queries failed: %s", stdout)
	}
	if !strings.Contains(stdout, "list_users") {
		t.Errorf("newly-init'd sibling does not see the shared group:\n%s", stdout)
	}
}

func TestRemoveConnection_KeepsGroupWhenSiblingRemains(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev":  env.NewDBPath("dev.db"),
		"ecommerce:prod": env.NewDBPath("prod.db"),
	}, "ecommerce:dev"))

	env.RunSquix("add", "list_users", "SELECT * FROM users")
	// Remove dev; prod remains and should keep the shared group.
	if _, _, code := env.RunSquixWithStdin("y\n", "remove", "-c", "ecommerce:dev"); code != 0 {
		t.Fatal("remove dev failed")
	}

	env.RunSquix("switch", "ecommerce:prod")
	stdout, _, code := env.RunSquix("list", "queries")
	if code != 0 {
		t.Fatalf("list queries on prod failed: %s", stdout)
	}
	if !strings.Contains(stdout, "list_users") {
		t.Errorf("removing a sibling dropped the shared group:\n%s", stdout)
	}
}

func TestRemoveConnection_GCOrphanGroup(t *testing.T) {
	env := Setup(t)
	env.WriteConfig(multiConnConfig(map[string]string{
		"ecommerce:dev": env.NewDBPath("dev.db"),
	}, "ecommerce:dev"))

	env.RunSquix("add", "list_users", "SELECT * FROM users")

	grpFile := filepath.Join(env.HomeDir, ".config", "squix", "queries", "ecommerce.sql")
	if _, err := os.Stat(grpFile); err != nil {
		t.Fatalf("group file not created after add: %v", err)
	}
	if _, _, code := env.RunSquixWithStdin("y\n", "remove", "-c", "ecommerce:dev"); code != 0 {
		t.Fatal("remove dev failed")
	}

	// The orphaned group file must be removed with its last connection.
	if _, err := os.Stat(grpFile); !os.IsNotExist(err) {
		t.Errorf("orphaned group file not removed after removing last member: %v", err)
	}
}

func TestMigration_LegacyConfigUpgrades(t *testing.T) {
	env := Setup(t)
	legacy := `current_connection: onlydb
connections:
  onlydb:
    name: onlydb
    db_type: sqlite
    conn_string: ` + env.NewDBPath("only.db") + `
    queries:
      legacy_q:
        name: legacy_q
        id: 1
        sql: SELECT 1
`
	env.WriteConfig(legacy)

	// Any saving command triggers load+migrate+save.
	if _, _, code := env.RunSquix("add", "new_q", "SELECT 2"); code != 0 {
		t.Fatal("add on legacy config failed")
	}

	// Legacy queries migrate to a per-group .sql file.
	grpFile := filepath.Join(env.HomeDir, ".config", "squix", "queries", "onlydb.sql")
	data, err := os.ReadFile(grpFile)
	if err != nil {
		t.Fatalf("legacy queries not migrated to %s: %v", grpFile, err)
	}
	if !strings.Contains(string(data), "legacy_q") {
		t.Errorf("legacy query lost during migration:\n%s", data)
	}

	// config.yaml must no longer carry inline per-connection queries.
	cfgData, err := os.ReadFile(filepath.Join(env.HomeDir, ".config", "squix", "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if strings.Contains(string(cfgData), "queries:") {
		t.Errorf("inline per-connection queries still present after migration:\n%s", cfgData)
	}
}
