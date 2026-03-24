// Package gservice implements the gRPC server handlers for GophKeeper authentication
// and encrypted storage synchronization, along with request interceptors.
package gservice

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
)

// contextKey is an unexported type for context keys used in this package.
type contextKey string

// userIDKey is the context key for the authenticated user's ID.
const userIDKey contextKey = "user_id"

// wrappedStream wraps grpc.ServerStream to carry an overridden context.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the overridden context stored in the wrappedStream.
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// AuthInterceptor is a gRPC unary server interceptor that validates JWT tokens.
// Register and Login endpoints are exempt from authentication.
func (s *Server) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for Register and Login
	if info.FullMethod == "/gophkeeper.v1.AuthService/Register" || info.FullMethod == "/gophkeeper.v1.AuthService/Login" {
		return handler(ctx, req)
	}

	newCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

// StreamAuthInterceptor is a gRPC stream server interceptor that validates JWT tokens
// for all streaming RPCs (e.g. Sync).
func (s *Server) StreamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	newCtx, err := s.authorize(ss.Context())
	if err != nil {
		return err
	}
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: newCtx})
}

// authorize extracts and validates the Bearer JWT from incoming gRPC metadata,
// returning a new context enriched with the authenticated user_id.
func (s *Server) authorize(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is missing")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is missing")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user_id missing in claims")
	}

	return context.WithValue(ctx, userIDKey, userID), nil
}
