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
}

// Repository defines the contract for data persistence.
type Repository interface {
	// CreateUser creates a new user in the storage.
	CreateUser(ctx context.Context, username, passwordHash string) (*User, error)
	// GetUserByUsername retrieves a user by their unique username.
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	
	// SyncRecords synchronizes records for a specific user.
	// It upserts incoming records and returns updates since lastRevision.
	SyncRecords(ctx context.Context, userID uuid.UUID, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error)
}
