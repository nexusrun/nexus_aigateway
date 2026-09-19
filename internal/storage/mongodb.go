package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DefaultMongoDatabase is the database used when neither the explicit Database
// config nor the connection string names one. LegacyMongoDatabase is the name
// used before the gateway was renamed.
const (
	DefaultMongoDatabase = "aigateway"
	LegacyMongoDatabase  = "gomodel"
)

// mongoStorage implements Storage for MongoDB
type mongoStorage struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB storage connection.
func NewMongoDB(ctx context.Context, cfg MongoDBConfig) (MongoDBStorage, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("MongoDB URL is required")
	}

	dbName := resolveMongoDatabase(cfg)

	// Create client options
	clientOpts := options.Client().ApplyURI(cfg.URL)

	// Connect to MongoDB
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	if dbName == DefaultMongoDatabase {
		dbName = resolveLegacyMongoDatabase(ctx, client, dbName)
	}

	// Get database reference
	database := client.Database(dbName)

	return &mongoStorage{
		client:   client,
		database: database,
	}, nil
}

// resolveLegacyMongoDatabase keeps an unconfigured gateway pointed at the
// database it was already using before the rename. Mongo has no cheap rename,
// so the old database is adopted in place rather than copied. Listing requires
// a privilege a locked-down deployment may withhold; that is not fatal, because
// failing to list simply leaves the default in force.
func resolveLegacyMongoDatabase(ctx context.Context, client *mongo.Client, dbName string) string {
	names, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		slog.Debug("cannot list MongoDB databases; keeping default", "database", dbName, "error", err)
		return dbName
	}
	// Mongo only reports databases holding data, so presence here means the
	// legacy database has content worth keeping.
	if slices.Contains(names, dbName) || !slices.Contains(names, LegacyMongoDatabase) {
		return dbName
	}
	slog.Warn("using pre-rename MongoDB database; set MONGODB_DATABASE to override",
		"database", LegacyMongoDatabase, "renamed_default", dbName)
	return LegacyMongoDatabase
}

// resolveMongoDatabase picks the database name for a connection: the explicit
// Database config wins, then the database named in the connection string path
// (the standard "mongodb://host/dbname" form), then DefaultMongoDatabase.
func resolveMongoDatabase(cfg MongoDBConfig) string {
	if cfg.Database != "" {
		return cfg.Database
	}
	if name := databaseNameFromURI(cfg.URL); name != "" {
		return name
	}
	return DefaultMongoDatabase
}

// databaseNameFromURI extracts the database name (the path component) from a
// MongoDB connection string. Returns "" when the URI is unparseable, names no
// database, or the path is not a single segment — MongoDB database names
// cannot contain '/', so a multi-segment path would fail at connect time.
func databaseNameFromURI(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	name := strings.Trim(u.Path, "/")
	if strings.Contains(name, "/") {
		return ""
	}
	return name
}

func (s *mongoStorage) Close() error {
	if s.client != nil {
		return s.client.Disconnect(context.Background())
	}
	return nil
}

// Database returns the underlying *mongo.Database for direct access
func (s *mongoStorage) Database() *mongo.Database {
	return s.database
}

// Client returns the underlying *mongo.Client for direct access
func (s *mongoStorage) Client() *mongo.Client {
	return s.client
}

// Ping verifies connectivity to MongoDB.
func (s *mongoStorage) Ping(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("mongodb client is not initialized")
	}
	return s.client.Ping(ctx, nil)
}
