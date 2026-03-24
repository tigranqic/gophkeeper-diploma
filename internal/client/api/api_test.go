package api

import (
	"context"
	"net"
	"testing"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

// fakeAuthService is a minimal in-process implementation of AuthServiceServer.
type fakeAuthService struct {
	pb.UnimplementedAuthServiceServer
}

func (f *fakeAuthService) Register(_ context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Token: "register-token-for-" + req.Username}, nil
}

func (f *fakeAuthService) Login(_ context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{Token: "login-token-for-" + req.Username}, nil
}

// fakeStorageService is a minimal in-process implementation of StorageServiceServer.
type fakeStorageService struct {
	pb.UnimplementedStorageServiceServer
}

func (f *fakeStorageService) Sync(req *pb.SyncRequest, stream pb.StorageService_SyncServer) error {
	return stream.Send(&pb.SyncResponse{
		Records:         req.Records,
		CurrentRevision: req.LastRevision + 1,
	})
}

// startBufconnServer starts an in-memory gRPC server and returns a dialer for it.
func startBufconnServer(t *testing.T) func(context.Context, string) (net.Conn, error) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterAuthServiceServer(srv, &fakeAuthService{})
	pb.RegisterStorageServiceServer(srv, &fakeStorageService{})

	go srv.Serve(lis)
	t.Cleanup(func() { srv.Stop() })

	return func(ctx context.Context, _ string) (net.Conn, error) {
		return lis.DialContext(ctx)
	}
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	dialer := startBufconnServer(t)
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return &Client{
		authClient:    pb.NewAuthServiceClient(conn),
		storageClient: pb.NewStorageServiceClient(conn),
	}
}

func TestClient_Register(t *testing.T) {
	c := newTestClient(t)
	token, err := c.Register(context.Background(), "alice", "hash")
	require.NoError(t, err)
	assert.Equal(t, "register-token-for-alice", token)
}

func TestClient_Login(t *testing.T) {
	c := newTestClient(t)
	token, err := c.Login(context.Background(), "bob", "hash")
	require.NoError(t, err)
	assert.Equal(t, "login-token-for-bob", token)
}

func TestClient_SetToken(t *testing.T) {
	c := newTestClient(t)
	c.SetToken("new-token")
	assert.Equal(t, "new-token", c.token)
}

func TestClient_Sync(t *testing.T) {
	c := newTestClient(t)
	c.SetToken("some-token")

	records := []*pb.EncryptedRecord{
		{Id: "rec-1", UpdatedAt: 100},
	}

	updates, rev, err := c.Sync(context.Background(), records, int64(5))
	require.NoError(t, err)
	assert.Equal(t, int64(6), rev)
	assert.Len(t, updates, 1)
	assert.Equal(t, "rec-1", updates[0].Id)
}
