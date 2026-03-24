package gservice

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/mocks"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testSecret = "test-jwt-secret"

func makeServer() *Server {
	return NewServer(new(mocks.MockRepository), testSecret, zap.NewNop())
}

func makeValidToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "00000000-0000-0000-0000-000000000001",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	t, _ := token.SignedString([]byte(secret))
	return t
}

// noopUnaryHandler is a dummy gRPC unary handler that always succeeds.
func noopUnaryHandler(ctx context.Context, req interface{}) (interface{}, error) {
	return "ok", nil
}

func TestAuthInterceptor_SkipRegister(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.AuthService/Register"}
	res, err := s.AuthInterceptor(context.Background(), nil, info, noopUnaryHandler)
	require.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestAuthInterceptor_SkipLogin(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.AuthService/Login"}
	res, err := s.AuthInterceptor(context.Background(), nil, info, noopUnaryHandler)
	require.NoError(t, err)
	assert.Equal(t, "ok", res)
}

func TestAuthInterceptor_MissingMetadata(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.StorageService/Sync"}
	_, err := s.AuthInterceptor(context.Background(), nil, info, noopUnaryHandler)
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_MissingToken(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.StorageService/Sync"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})
	_, err := s.AuthInterceptor(ctx, nil, info, noopUnaryHandler)
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_InvalidToken(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.StorageService/Sync"}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer invalid-token"))
	_, err := s.AuthInterceptor(ctx, nil, info, noopUnaryHandler)
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestAuthInterceptor_ValidToken(t *testing.T) {
	s := makeServer()
	info := &grpc.UnaryServerInfo{FullMethod: "/gophkeeper.v1.StorageService/Sync"}
	token := makeValidToken(testSecret)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	var capturedCtx context.Context
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		capturedCtx = ctx
		return "ok", nil
	}

	res, err := s.AuthInterceptor(ctx, nil, info, handler)
	require.NoError(t, err)
	assert.Equal(t, "ok", res)
	assert.NotNil(t, capturedCtx.Value(userIDKey))
}

func TestStreamAuthInterceptor_MissingMetadata(t *testing.T) {
	s := makeServer()
	stream := &mockSyncServer{ctx: context.Background()}
	err := s.StreamAuthInterceptor(nil, stream, nil, func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestStreamAuthInterceptor_ValidToken(t *testing.T) {
	s := makeServer()
	token := makeValidToken(testSecret)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
	stream := &mockSyncServer{ctx: ctx}

	var capturedCtx context.Context
	err := s.StreamAuthInterceptor(nil, stream, nil, func(srv interface{}, ss grpc.ServerStream) error {
		capturedCtx = ss.Context()
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, capturedCtx.Value(userIDKey))
}
