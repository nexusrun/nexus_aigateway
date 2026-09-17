package adminauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoAccountDocument struct {
	ID           string    `bson:"_id"`
	Username     string    `bson:"username"`
	PasswordHash string    `bson:"password_hash"`
	Role         string    `bson:"role"`
	Enabled      bool      `bson:"enabled"`
	CreatedAt    time.Time `bson:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at"`
}

type mongoStore struct {
	collection *mongo.Collection
}

func newMongoStore(ctx context.Context, database *mongo.Database) (Store, error) {
	if database == nil {
		return nil, fmt.Errorf("database is required")
	}
	collection := database.Collection("admin_accounts")
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("create admin_accounts username index: %w", err)
	}
	return &mongoStore{collection: collection}, nil
}

func (s *mongoStore) FindByUsername(ctx context.Context, username string) (Account, error) {
	return s.findOne(ctx, bson.D{{Key: "username", Value: username}})
}

func (s *mongoStore) FindByID(ctx context.Context, id string) (Account, error) {
	return s.findOne(ctx, bson.D{{Key: "_id", Value: id}})
}

func (s *mongoStore) findOne(ctx context.Context, filter any) (Account, error) {
	var document mongoAccountDocument
	if err := s.collection.FindOne(ctx, filter).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Account{}, ErrNotFound
		}
		return Account{}, fmt.Errorf("find admin account: %w", err)
	}
	return accountFromMongo(document), nil
}

func (s *mongoStore) Create(ctx context.Context, account Account) error {
	_, err := s.collection.InsertOne(ctx, mongoAccountDocument{
		ID:           account.ID,
		Username:     account.Username,
		PasswordHash: account.PasswordHash,
		Role:         account.Role,
		Enabled:      account.Enabled,
		CreatedAt:    account.CreatedAt.UTC(),
		UpdatedAt:    account.UpdatedAt.UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create admin account: %w", err)
	}
	return nil
}

func (s *mongoStore) Close() error { return nil }

func accountFromMongo(document mongoAccountDocument) Account {
	return Account{
		ID:           document.ID,
		Username:     document.Username,
		PasswordHash: document.PasswordHash,
		Role:         document.Role,
		Enabled:      document.Enabled,
		CreatedAt:    document.CreatedAt.UTC(),
		UpdatedAt:    document.UpdatedAt.UTC(),
	}
}
