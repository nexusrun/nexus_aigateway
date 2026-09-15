package adminauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nexusrun/nexus_aigateway/internal/storage/sqlx"
)

type sqlStore struct{ db sqlx.DB }

const accountTable = `CREATE TABLE IF NOT EXISTS admin_accounts (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'admin',
	enabled ` + sqlx.TypeBool + ` NOT NULL DEFAULT TRUE,
	created_at ` + sqlx.TypeInt64 + ` NOT NULL,
	updated_at ` + sqlx.TypeInt64 + ` NOT NULL
)`

func newSQLStore(ctx context.Context, db sqlx.DB) (Store, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}
	if err := db.Schema(ctx, accountTable); err != nil {
		return nil, fmt.Errorf("create admin_accounts table: %w", err)
	}
	return &sqlStore{db: db}, nil
}

const accountColumns = `SELECT id, username, password_hash, role, enabled, created_at, updated_at FROM admin_accounts `

func scanAccount(scanner interface{ Scan(...any) error }) (Account, error) {
	var a Account
	var created, updated int64
	if err := scanner.Scan(&a.ID, &a.Username, &a.PasswordHash, &a.Role, &a.Enabled, &created, &updated); err != nil {
		return Account{}, err
	}
	a.CreatedAt = time.Unix(created, 0).UTC()
	a.UpdatedAt = time.Unix(updated, 0).UTC()
	return a, nil
}

func (s *sqlStore) FindByUsername(ctx context.Context, username string) (Account, error) {
	row := s.db.QueryRow(ctx, accountColumns+`WHERE username = ?`, username)
	a, err := scanAccount(row)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) { return Account{}, ErrNotFound }
		return Account{}, fmt.Errorf("find admin account: %w", err)
	}
	return a, nil
}

func (s *sqlStore) FindByID(ctx context.Context, id string) (Account, error) {
	row := s.db.QueryRow(ctx, accountColumns+`WHERE id = ?`, id)
	a, err := scanAccount(row)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) { return Account{}, ErrNotFound }
		return Account{}, fmt.Errorf("find admin account: %w", err)
	}
	return a, nil
}

func (s *sqlStore) Create(ctx context.Context, a Account) error {
	_, err := s.db.Exec(ctx, `INSERT INTO admin_accounts (id, username, password_hash, role, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, a.ID, a.Username, a.PasswordHash, a.Role, a.Enabled, a.CreatedAt.Unix(), a.UpdatedAt.Unix())
	if err != nil { return fmt.Errorf("create admin account: %w", err) }
	return nil
}

func (s *sqlStore) Close() error { return nil }
