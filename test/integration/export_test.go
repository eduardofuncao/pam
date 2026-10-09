package integration

import (
	"strings"
	"testing"
)

func TestExportUsageWithoutSubcommand(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	_, stderr, exitCode := env.RunSquix("export")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(stderr, "usage: squix export [backup|schema]") {
		t.Errorf("expected usage message:\n%s", stderr)
	}
}

func TestExportUnsupportedEngine(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	_, stderr, exitCode := env.RunSquix("export", "schema")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(stderr, "native backup not implemented for sqlite") {
		t.Errorf("expected unsupported engine error:\n%s", stderr)
	}
}

func TestExportTableFlagUnderBackup(t *testing.T) {
	env := Setup(t)
	env.SeedDefaults()

	_, stderr, exitCode := env.RunSquix("export", "backup", "--table", "users")
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit")
	}
	if !strings.Contains(stderr, "--table is only valid for export schema") {
		t.Errorf("expected --table error:\n%s", stderr)
	}
}
