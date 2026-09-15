package adminauth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("admin account not found")
	ErrInvalidLogin   = errors.New("invalid username or password")
	ErrAlreadyExists  = errors.New("admin account already exists")
)

type Account struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Store interface {
	FindByUsername(context.Context, string) (Account, error)
	FindByID(context.Context, string) (Account, error)
	Create(context.Context, Account) error
	Close() error
}
