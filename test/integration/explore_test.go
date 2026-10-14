package integration

import (
	"strings"
	"testing"
)

func TestExploreOnelineListsTables(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	stdout, _, exitCode := env.RunSquix("explore", "--oneline")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(stdout, "users") || !strings.Contains(stdout, "orders") {
		t.Errorf("expected users and orders:\n%s", stdout)
	}
}

func TestExploreOnelineTablesFlag(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	stdout, _, exitCode := env.RunSquix("explore", "--oneline", "--tables")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(stdout, "users") {
		t.Errorf("expected users:\n%s", stdout)
	}
}

func TestExploreOnelineViewsFlag(t *testing.T) {
	env := Setup(t)
	env.SeedData(
		DefaultSchema+"CREATE VIEW adult_users AS SELECT id, name FROM users WHERE age >= 18;",
		DefaultInserts,
	)

	tablesOut, _, exitCode := env.RunSquix("explore", "--oneline", "--tables")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if strings.Contains(tablesOut, "adult_users") {
		t.Errorf("--tables should not list views:\n%s", tablesOut)
	}

	viewsOut, _, exitCode := env.RunSquix("explore", "--oneline", "--views")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(viewsOut, "adult_users") {
		t.Errorf("--views should list the view:\n%s", viewsOut)
	}
	if strings.Contains(viewsOut, "orders") {
		t.Errorf("--views should not list base tables:\n%s", viewsOut)
	}
}

func TestTablesCommandRemoved(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	_, _, exitCode := env.RunSquix("tables")
	if exitCode == 0 {
		t.Errorf("squix tables should no longer exist")
	}
}

func TestExplainTable(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	stdout, _, exitCode := env.RunSquix("explain", "users")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(stdout, "users") {
		t.Errorf("expected users in explain output:\n%s", stdout)
	}
}

func TestExplainWithDepth(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	stdout, _, exitCode := env.RunSquix("explain", "users", "-d", "2")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(stdout, "orders") {
		t.Errorf("expected FK relationship to orders:\n%s", stdout)
	}
}

func TestExplainVerbose(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	stdout, _, exitCode := env.RunSquix("explain", "users", "-v")
	if exitCode != 0 {
		t.Fatalf("exit %d", exitCode)
	}
	if !strings.Contains(stdout, "users") {
		t.Errorf("expected verbose explain:\n%s", stdout)
	}
}
