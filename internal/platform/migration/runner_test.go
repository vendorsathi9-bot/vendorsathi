package migration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMigrationFile(t *testing.T) {
	m, err := parseMigrationFile("/tmp/001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("parse valid migration: %v", err)
	}
	if m.Version != 1 || m.Name != "initial_schema" {
		t.Fatalf("unexpected migration: %+v", m)
	}

	if _, err := parseMigrationFile("/tmp/not-versioned.up.sql"); err == nil {
		t.Fatal("expected error for filename without a numeric version")
	}
}

func TestFilterPending(t *testing.T) {
	all := []Migration{
		{Version: 1, Name: "one"},
		{Version: 2, Name: "two"},
		{Version: 3, Name: "three"},
	}
	applied := map[int64]struct{}{1: {}, 3: {}}

	pending := filterPending(all, applied)
	if len(pending) != 1 || pending[0].Version != 2 {
		t.Fatalf("expected only migration 2 to be pending, got %+v", pending)
	}
}

func TestDiscoverMigrations(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"001_initial_schema.up.sql": "SELECT 1;",
		"002_vendor.up.sql":         "SELECT 2;",
		"002_vendor.down.sql":       "SELECT 3;",
		"README.md":                 "ignored",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	r := &Runner{migrDir: dir}
	migrations, err := r.discoverMigrations()
	if err != nil {
		t.Fatalf("discover migrations: %v", err)
	}
	if len(migrations) != 2 {
		t.Fatalf("expected 2 up migrations, got %d: %+v", len(migrations), migrations)
	}
	if migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Fatalf("migrations not sorted by version: %+v", migrations)
	}
}
