// Package mocks provides testify mock implementations of repository interfaces.
package mocks

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
)

// MockRepository is a testify mock that implements repository.Repository.
type MockRepository struct {
	mock.Mock
}

// CreateUser records the call and returns the configured stub values.
func (m *MockRepository) CreateUser(ctx context.Context, username, passwordHash string) (*repository.User, error) {
	args := m.Called(ctx, username, passwordHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

// GetUserByUsername records the call and returns the configured stub values.
func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

// SyncRecords records the call and returns the configured stub values.
func (m *MockRepository) SyncRecords(ctx context.Context, userID uuid.UUID, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error) {
	args := m.Called(ctx, userID, records, lastRevision)
	return args.Get(0).([]*pb.EncryptedRecord), args.Get(1).(int64), args.Error(2)
}
