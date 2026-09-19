package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// sqliteStorage implements Storage for SQLite
type sqliteStorage struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite storage connection.
// It enables WAL mode for better concurrent read/write performance.
func NewSQLite(cfg SQLiteConfig) (SQLiteStorage, error) {
	if cfg.Path == "" {
		cfg.Path = DefaultSQLitePath()
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := adoptLegacySQLiteFile(cfg.Path); err != nil {
		return nil, err
	}

	// Open database with WAL mode and busy timeout
	dsn := fmt.Sprintf("%s?_journal=WAL&_busy_timeout=5000&_synchronous=NORMAL", cfg.Path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Serialize all SQLite access through a single connection.
	// SQLite only permits one writer at a time. With multiple open connections,
	// concurrent flush loops (audit log + usage tracking) cause SQLITE_BUSY errors.
	// A single connection lets database/sql queue callers in Go, eliminating contention.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Set busy_timeout explicitly as defense-in-depth. The DSN parameter may not be
	// honored reliably by the pure-Go driver across pooled connections.
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy_timeout pragma: %w", err)
	}

	return &sqliteStorage{db: db}, nil
}

// adoptLegacySQLiteFile renames a database still stored under
// LegacySQLiteFilename to path, so a gateway upgraded past the rename keeps its
// data. It does nothing when path is not the current default filename, or when
// a database already exists there.
func adoptLegacySQLiteFile(path string) error {
	if filepath.Base(path) != SQLiteFilename {
		return nil
	}
	legacy := filepath.Join(filepath.Dir(path), LegacySQLiteFilename)
	if !fileExists(legacy) || fileExists(path) {
		return nil
	}

	// The side files move first: interrupted after the database alone is
	// renamed, the next start sees the new name and skips adoption, orphaning a
	// -wal that still holds committed transactions. Interrupted in this order,
	// the database keeps its old name and adoption simply resumes.
	for _, suffix := range []string{"-wal", "-shm", ""} {
		if !fileExists(legacy + suffix) {
			continue
		}
		if err := os.Rename(legacy+suffix, path+suffix); err != nil {
			return fmt.Errorf("failed to rename legacy database %s: %w", legacy+suffix, err)
		}
	}
	slog.Info("adopted legacy SQLite database", "from", legacy, "to", path)
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (s *sqliteStorage) DB() *sql.DB {
	return s.db
}

// Ping verifies connectivity to the SQLite database.
func (s *sqliteStorage) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("sqlite database is not initialized")
	}
	return s.db.PingContext(ctx)
}

func (s *sqliteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
