package storage

import (
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoTransactionFallbackError marks an error raised inside a transaction
// callback so the caller can retry the same work without a transaction.
type MongoTransactionFallbackError struct {
	err error
}

// NewMongoTransactionFallbackError wraps err for MongoTransactionFallbackCause.
func NewMongoTransactionFallbackError(err error) *MongoTransactionFallbackError {
	return &MongoTransactionFallbackError{err: err}
}

func (e *MongoTransactionFallbackError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

// MongoTransactionFallbackCause returns the wrapped cause when err came from a
// transaction callback that flagged itself for a non-transactional retry.
func MongoTransactionFallbackCause(err error) error {
	if fallbackErr, ok := errors.AsType[*MongoTransactionFallbackError](err); ok {
		return fallbackErr.err
	}
	return nil
}

// IsMongoTransactionCapabilityError reports whether err means the server cannot
// run transactions at all — a standalone mongod rather than a replica set or
// mongos — rather than a transaction that failed on its own merits.
func IsMongoTransactionCapabilityError(err error) bool {
	if err == nil {
		return false
	}
	var commandErr mongo.CommandError
	if errors.As(err, &commandErr) && commandErr.HasErrorCode(20) {
		return true
	}
	var labeled mongo.LabeledError
	if errors.As(err, &labeled) && labeled.HasErrorLabel("TransientTransactionError") {
		message := strings.ToLower(err.Error())
		return strings.Contains(message, "transaction") &&
			(strings.Contains(message, "not supported") ||
				strings.Contains(message, "not allowed") ||
				strings.Contains(message, "replica set"))
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "transaction numbers are only allowed on a replica set member or mongos")
}
