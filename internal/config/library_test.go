package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/eduardofuncao/squix/internal/db"
)

func TestQueryGroupKey(t *testing.T) {
	cases := []struct{ name, want string }{
		{"ecommerce:dev", "ecommerce"},
		{"ecommerce:prod", "ecommerce"},
		{"ecommerce:stg", "ecommerce"},
		{"mydb", "mydb"},
		{"", ""},
		{":dev", ":dev"}, // empty service → whole name (private)
		{"dev:", "dev"},  // empty env → service key "dev"
		{"ecommerce:us:dev", "ecommerce:us"},
	}
	for _, c := range cases {
		got := QueryGroupKey(c.name)
		if got != c.want {
			t.Errorf("QueryGroupKey(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestQueriesFor_SharedMutation(t *testing.T) {
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev":  {Name: "ecommerce:dev"},
			"ecommerce:prod": {Name: "ecommerce:prod"},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}

	dev := cfg.QueriesFor("ecommerce:dev")
	dev["list_users"] = db.Query{Name: "list_users", Id: 1, SQL: "SELECT 1"}

	prod := cfg.QueriesFor("ecommerce:prod")
	if _, ok := prod["list_users"]; !ok {
		t.Fatal("prod does not see query added via dev")
	}
	// Same underlying map reference → mutations propagate.
	if reflect.ValueOf(dev).Pointer() != reflect.ValueOf(prod).Pointer() {
		t.Fatal("dev and prod queries maps are not the same reference")
	}
}

func TestQueriesFor_CreatesLibrary(t *testing.T) {
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{"analytics": {Name: "analytics"}},
		QueryGroups: map[string]map[string]db.Query{},
	}
	q := cfg.QueriesFor("analytics")
	if q == nil {
		t.Fatal("QueriesFor returned nil")
	}
	q["x"] = db.Query{Name: "x"}
	if cfg.QueryGroups["analytics"]["x"].Name != "x" {
		t.Fatal("QueriesFor did not store the created library under the right key")
	}
}

func TestSetQueriesFor_PreservesReference(t *testing.T) {
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev":  {Name: "ecommerce:dev"},
			"ecommerce:prod": {Name: "ecommerce:prod"},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}
	dev := cfg.QueriesFor("ecommerce:dev")

	cfg.SetQueriesFor("ecommerce:dev", map[string]db.Query{
		"q1": {Name: "q1", Id: 1},
		"q2": {Name: "q2", Id: 2},
	})

	prod := cfg.QueriesFor("ecommerce:prod")
	if len(prod) != 2 {
		t.Fatalf("prod has %d queries after SetQueriesFor, want 2", len(prod))
	}
	// Reference preserved: dev still points at the same (now repopulated) map.
	if reflect.ValueOf(dev).Pointer() != reflect.ValueOf(prod).Pointer() {
		t.Fatal("SetQueriesFor swapped the map reference instead of replacing in place")
	}
}

func TestMigrateQueriesToFiles_LegacyInline(t *testing.T) {
	withTempCfg(t)
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev": {
				Name:    "ecommerce:dev",
				Queries: map[string]db.Query{"list_users": {Name: "list_users", SQL: "SELECT 1"}},
			},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}
	if !cfg.MigrateQueriesToFiles() {
		t.Fatal("expected migration to report it migrated something")
	}

	if cfg.Connections["ecommerce:dev"].Queries != nil {
		t.Error("legacy inline Queries not cleared after migration")
	}
	if got := cfg.QueryGroups["ecommerce"]["list_users"]; got.Name != "list_users" {
		t.Errorf("legacy query not lifted into group: %+v", got)
	}
	data, err := os.ReadFile(GroupFile("ecommerce"))
	if err != nil {
		t.Fatalf("group file not written: %v", err)
	}
	if !strings.Contains(string(data), "list_users") {
		t.Errorf("group file missing query:\n%s", data)
	}
}

func TestMigrateQueriesToFiles_NoColon(t *testing.T) {
	withTempCfg(t)
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"mydb": {Queries: map[string]db.Query{"q": {Name: "q"}}},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}
	cfg.MigrateQueriesToFiles()
	if cfg.QueryGroups["mydb"]["q"].Name != "q" {
		t.Error("no-colon connection did not map to a private group keyed by its name")
	}
	if _, err := os.Stat(GroupFile("mydb")); err != nil {
		t.Errorf("expected queries/mydb.sql: %v", err)
	}
}

func TestMigrateQueriesToFiles_FirstWinsSorted(t *testing.T) {
	withTempCfg(t)
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:prod": {Queries: map[string]db.Query{"prod_q": {Name: "prod_q"}}},
			"ecommerce:dev":  {Queries: map[string]db.Query{"dev_q": {Name: "dev_q"}}},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}
	cfg.MigrateQueriesToFiles()

	// Sorted: "ecommerce:dev" < "ecommerce:prod" → dev wins.
	if _, ok := cfg.QueryGroups["ecommerce"]["dev_q"]; !ok {
		t.Error("first-wins: expected dev_q in the group")
	}
	if _, ok := cfg.QueryGroups["ecommerce"]["prod_q"]; ok {
		t.Error("first-wins: prod_q should have been dropped")
	}
	data, _ := os.ReadFile(GroupFile("ecommerce"))
	if strings.Contains(string(data), "prod_q") {
		t.Errorf("dropped query leaked into file:\n%s", data)
	}
}

func TestMigrateQueriesToFiles_NilSafe(t *testing.T) {
	withTempCfg(t)
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{"x": {Queries: map[string]db.Query{"q": {Name: "q"}}}},
	}
	cfg.MigrateQueriesToFiles() // QueryGroups is nil here
	if cfg.QueryGroups == nil {
		t.Fatal("MigrateQueriesToFiles did not initialize QueryGroups")
	}
}

func TestMigrateQueriesToFiles_Idempotent(t *testing.T) {
	withTempCfg(t)
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev": {Queries: map[string]db.Query{"q": {Name: "q"}}},
		},
		QueryGroups: map[string]map[string]db.Query{},
	}
	if !cfg.MigrateQueriesToFiles() {
		t.Fatal("first migration should report migrated=true")
	}
	before := cfg.QueryGroups["ecommerce"]["q"].Name
	if cfg.MigrateQueriesToFiles() {
		t.Error("second migration should be a no-op (migrated=false)")
	}
	if cfg.QueryGroups["ecommerce"]["q"].Name != before {
		t.Error("second migration pass altered already-migrated data")
	}
}

func TestGroupHasOtherMembers(t *testing.T) {
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev":  {},
			"ecommerce:prod": {},
			"analytics":      {},
		},
	}
	if !cfg.GroupHasOtherMembers("ecommerce", "ecommerce:dev") {
		t.Error("expected a sibling sharing the ecommerce group")
	}
	if cfg.GroupHasOtherMembers("analytics", "analytics") {
		t.Error("standalone connection should have no siblings")
	}
}

func TestLiveConnection_PopulatesQueries(t *testing.T) {
	cfg := &Config{
		Connections: map[string]*ConnectionYAML{
			"ecommerce:dev": {Name: "ecommerce:dev", DBType: "sqlite", ConnString: ":memory:"},
		},
		QueryGroups: map[string]map[string]db.Query{
			"ecommerce": {"list_users": {Name: "list_users", Id: 1, SQL: "SELECT 1"}},
		},
	}
	conn := cfg.LiveConnection("ecommerce:dev")
	qs := conn.GetQueries()
	if len(qs) != 1 || qs["list_users"].Name != "list_users" {
		t.Errorf("LiveConnection did not populate queries from the shared group: %+v", qs)
	}
}
