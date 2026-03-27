// Package repository defines the interfaces and data models for server-side storage.
package repository

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
)

// User represents a system user with authentication credentials.
type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Salt         []byte
}

// UserRepository handles persistence of user accounts.
// Used exclusively by the authentication handlers (Register, Login).
type UserRepository interface {
	// CreateUser creates a new user in the storage.
	CreateUser(ctx context.Context, username, passwordHash string, salt []byte) (*User, error)
	// GetUserByUsername retrieves a user by their unique username.
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	// GetSaltByUsername returns the scrypt salt for the given username.
	// Returns (nil, nil) when no matching user exists.
	GetSaltByUsername(ctx context.Context, username string) ([]byte, error)
}

// RecordRepository handles persistence of encrypted records.
// Used exclusively by the storage handler (Sync).
type RecordRepository interface {
	// SyncRecords upserts incoming records and returns all updates since lastRevision.
	SyncRecords(ctx context.Context, userID uuid.UUID, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error)
}
