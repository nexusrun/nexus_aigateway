package storage

import (
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func TestIsMongoTransactionCapabilityError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "standalone transaction message",
			err:  errors.New("Transaction numbers are only allowed on a replica set member or mongos"),
			want: true,
		},
		{
			name: "illegal operation command code",
			err: mongo.CommandError{
				Code:    20,
				Message: "transaction is not supported by this deployment",
				Labels:  []string{"TransientTransactionError"},
			},
			want: true,
		},
		{
			name: "labeled unsupported transaction message",
			err: mongo.CommandError{
				Message: "transaction is not supported by this deployment",
				Labels:  []string{"TransientTransactionError"},
			},
			want: true,
		},
		{
			name: "ordinary transient transaction error",
			err: mongo.CommandError{
				Message: "temporary write conflict",
				Labels:  []string{"TransientTransactionError"},
			},
			want: false,
		},
		{
			name: "ordinary error",
			err:  errors.New("network timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMongoTransactionCapabilityError(tt.err); got != tt.want {
				t.Fatalf("IsMongoTransactionCapabilityError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestMongoTransactionFallbackCause(t *testing.T) {
	cause := errors.New("standalone mongod")
	if got := MongoTransactionFallbackCause(NewMongoTransactionFallbackError(cause)); got != cause {
		t.Fatalf("MongoTransactionFallbackCause = %v, want %v", got, cause)
	}
	if got := MongoTransactionFallbackCause(errors.New("other")); got != nil {
		t.Fatalf("MongoTransactionFallbackCause = %v, want nil", got)
	}
}
