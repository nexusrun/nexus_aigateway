package adminauth

import (
	"context"
	"testing"
	"time"

	"github.com/nexusrun/nexus_aigateway/internal/storage/mongotest"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestMongoStoreAccountRoundTrip(t *testing.T) {
	mongotest.Run(t, func(t *testing.T, database *mongo.Database) {
		ctx := context.Background()
		store, err := newMongoStore(ctx, database)
		require.NoError(t, err)

		now := time.Date(2026, time.September, 17, 21, 30, 0, 0, time.UTC)
		account := Account{
			ID:           "admin-id",
			Username:     "admin@nexusai.io",
			PasswordHash: "$2a$10$test",
			Role:         "admin",
			Enabled:      true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		require.NoError(t, store.Create(ctx, account))

		byUsername, err := store.FindByUsername(ctx, account.Username)
		require.NoError(t, err)
		require.Equal(t, account, byUsername)

		byID, err := store.FindByID(ctx, account.ID)
		require.NoError(t, err)
		require.Equal(t, account, byID)

		_, err = store.FindByUsername(ctx, "missing@nexusai.io")
		require.ErrorIs(t, err, ErrNotFound)
	})
}

func TestMongoStoreRejectsDuplicateUsername(t *testing.T) {
	mongotest.Run(t, func(t *testing.T, database *mongo.Database) {
		ctx := context.Background()
		store, err := newMongoStore(ctx, database)
		require.NoError(t, err)

		account := Account{
			ID:           "admin-one",
			Username:     "admin@nexusai.io",
			PasswordHash: "$2a$10$test",
			Role:         "admin",
			Enabled:      true,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}
		require.NoError(t, store.Create(ctx, account))

		account.ID = "admin-two"
		require.ErrorIs(t, store.Create(ctx, account), ErrAlreadyExists)
	})
}
