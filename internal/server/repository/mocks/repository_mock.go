// Package mocks provides testify mock implementations of repository interfaces.
package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
)

// MockUserRepository is a testify mock that implements repository.UserRepository.
type MockUserRepository struct {
	mock.Mock
}

// CreateUser records the call and returns the configured stub values.
func (m *MockUserRepository) CreateUser(ctx context.Context, username, passwordHash string, salt []byte) (*repository.User, error) {
	args := m.Called(ctx, username, passwordHash, salt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

// GetSaltByUsername records the call and returns the configured stub values.
func (m *MockUserRepository) GetSaltByUsername(ctx context.Context, username string) ([]byte, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// GetUserByUsername records the call and returns the configured stub values.
func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

// MockRecordRepository is a testify mock that implements repository.RecordRepository.
type MockRecordRepository struct {
	mock.Mock
}

// SyncRecords records the call and returns the configured stub values.
func (m *MockRecordRepository) SyncRecords(ctx context.Context, userID uuid.UUID, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error) {
	args := m.Called(ctx, userID, records, lastRevision)
	return args.Get(0).([]*pb.EncryptedRecord), args.Get(1).(int64), args.Error(2)
}
