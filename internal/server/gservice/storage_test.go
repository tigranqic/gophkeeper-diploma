package gservice

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"testing"
)

// mockSyncServer implements pb.StorageService_SyncServer for testing.
type mockSyncServer struct {
	grpc.ServerStream
	ctx     context.Context
	sent    []*pb.SyncResponse
	sendErr error
}

func (m *mockSyncServer) Context() context.Context {
	return m.ctx
}

func (m *mockSyncServer) Send(res *pb.SyncResponse) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, res)
	return nil
}

func TestStorageService_Sync_Unauthenticated(t *testing.T) {
	mockUsers := new(mocks.MockUserRepository)
	mockRecords := new(mocks.MockRecordRepository)
	server := NewServer(mockUsers, mockRecords, WithJWTSecret("test-secret"), WithLogger(zap.NewNop()))

	// Context without user_id → should return Unauthenticated.
	stream := &mockSyncServer{ctx: context.Background()}
	err := server.Sync(&pb.SyncRequest{}, stream)
	require.Error(t, err)
}

func TestStorageService_Sync_EmptyUpdates(t *testing.T) {
	mockUsers := new(mocks.MockUserRepository)
	mockRecords := new(mocks.MockRecordRepository)
	server := NewServer(mockUsers, mockRecords, WithJWTSecret("test-secret"), WithLogger(zap.NewNop()))

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDKey, userID.String())

	mockRecords.On("SyncRecords", mock.Anything, userID, mock.Anything, int64(0)).
		Return([]*pb.EncryptedRecord{}, int64(7), nil)

	stream := &mockSyncServer{ctx: ctx}
	err := server.Sync(&pb.SyncRequest{LastRevision: 0}, stream)
	require.NoError(t, err)
	require.Len(t, stream.sent, 1)
	assert.Equal(t, int64(7), stream.sent[0].CurrentRevision)
	mockRecords.AssertExpectations(t)
}

func TestStorageService_Sync(t *testing.T) {
	mockUsers := new(mocks.MockUserRepository)
	mockRecords := new(mocks.MockRecordRepository)
	server := NewServer(mockUsers, mockRecords, WithJWTSecret("test-secret"), WithLogger(zap.NewNop()))

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), userIDKey, userID.String())

	records := []*pb.EncryptedRecord{
		{Id: "1", UpdatedAt: 100},
	}
	lastRevision := int64(5)

	mockRecords.On("SyncRecords", mock.Anything, userID, records, lastRevision).Return([]*pb.EncryptedRecord{
		{Id: "2", UpdatedAt: 150, Revision: 10},
	}, int64(10), nil)

	req := &pb.SyncRequest{
		Records:      records,
		LastRevision: lastRevision,
	}

	stream := &mockSyncServer{ctx: ctx}
	err := server.Sync(req, stream)

	require.NoError(t, err)
	assert.Len(t, stream.sent, 1)
	assert.Equal(t, int64(10), stream.sent[0].CurrentRevision)
	assert.Equal(t, "2", stream.sent[0].Records[0].Id)
	mockRecords.AssertExpectations(t)
}
