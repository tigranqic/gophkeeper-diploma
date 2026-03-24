package mocks

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
)

func TestMockRepository_CreateUser(t *testing.T) {
	m := new(MockRepository)
	expected := &repository.User{ID: uuid.New(), Username: "alice"}
	m.On("CreateUser", mock.Anything, "alice", "hash").Return(expected, nil)

	got, err := m.CreateUser(context.Background(), "alice", "hash")
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	m.AssertExpectations(t)
}

func TestMockRepository_GetUserByUsername(t *testing.T) {
	m := new(MockRepository)
	expected := &repository.User{ID: uuid.New(), Username: "bob"}
	m.On("GetUserByUsername", mock.Anything, "bob").Return(expected, nil)

	got, err := m.GetUserByUsername(context.Background(), "bob")
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	m.AssertExpectations(t)
}

func TestMockRepository_SyncRecords(t *testing.T) {
	m := new(MockRepository)
	userID := uuid.New()
	records := []*pb.EncryptedRecord{{Id: "1"}}
	expected := []*pb.EncryptedRecord{{Id: "2"}}

	m.On("SyncRecords", mock.Anything, userID, records, int64(0)).
		Return(expected, int64(5), nil)

	got, rev, err := m.SyncRecords(context.Background(), userID, records, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), rev)
	assert.Equal(t, expected, got)
	m.AssertExpectations(t)
}
