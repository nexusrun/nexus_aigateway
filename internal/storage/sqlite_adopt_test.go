package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAdoptLegacySQLiteFile(t *testing.T) {
	t.Run("renames database and its side files", func(t *testing.T) {
		dir := t.TempDir()
		legacy := filepath.Join(dir, LegacySQLiteFilename)
		target := filepath.Join(dir, SQLiteFilename)
		writeFile(t, legacy, "db")
		writeFile(t, legacy+"-wal", "wal")
		writeFile(t, legacy+"-shm", "shm")

		if err := adoptLegacySQLiteFile(target); err != nil {
			t.Fatalf("adoptLegacySQLiteFile() error: %v", err)
		}

		for suffix, want := range map[string]string{"": "db", "-wal": "wal", "-shm": "shm"} {
			got, err := os.ReadFile(target + suffix)
			if err != nil {
				t.Fatalf("reading %s: %v", target+suffix, err)
			}
			if string(got) != want {
				t.Errorf("%s = %q, want %q", target+suffix, got, want)
			}
			if fileExists(legacy + suffix) {
				t.Errorf("%s still exists", legacy+suffix)
			}
		}
	})

	t.Run("keeps an existing database untouched", func(t *testing.T) {
		dir := t.TempDir()
		legacy := filepath.Join(dir, LegacySQLiteFilename)
		target := filepath.Join(dir, SQLiteFilename)
		writeFile(t, legacy, "legacy")
		writeFile(t, target, "current")

		if err := adoptLegacySQLiteFile(target); err != nil {
			t.Fatalf("adoptLegacySQLiteFile() error: %v", err)
		}

		got, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "current" {
			t.Errorf("target = %q, want %q", got, "current")
		}
		if !fileExists(legacy) {
			t.Error("legacy database was removed")
		}
	})

	t.Run("ignores a non-default filename", func(t *testing.T) {
		dir := t.TempDir()
		legacy := filepath.Join(dir, LegacySQLiteFilename)
		writeFile(t, legacy, "legacy")
		target := filepath.Join(dir, "custom.db")

		if err := adoptLegacySQLiteFile(target); err != nil {
			t.Fatalf("adoptLegacySQLiteFile() error: %v", err)
		}

		if fileExists(target) {
			t.Error("custom.db should not have been created")
		}
		if !fileExists(legacy) {
			t.Error("legacy database was moved")
		}
	})

	t.Run("no legacy database is a no-op", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, SQLiteFilename)

		if err := adoptLegacySQLiteFile(target); err != nil {
			t.Fatalf("adoptLegacySQLiteFile() error: %v", err)
		}
		if fileExists(target) {
			t.Error("target should not have been created")
		}
	})
}
