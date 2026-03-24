package gservice

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/postgres"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// Server implements the AuthService and StorageService gRPC interfaces.
type Server struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedStorageServiceServer
	repo      repository.Repository
	jwtSecret string
	log       *zap.Logger
}

// NewServer creates a new Server backed by the given repository, signed with jwtSecret.
func NewServer(repo repository.Repository, jwtSecret string, log *zap.Logger) *Server {
	return &Server{repo: repo, jwtSecret: jwtSecret, log: log}
}

// Register creates a new user account. The incoming password hash is hashed again
// with bcrypt before storage. Returns a signed JWT token on success.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.PasswordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to process request")
	}

	user, err := s.repo.CreateUser(ctx, req.Username, string(hashed))
	if err != nil {
		if errors.Is(err, postgres.ErrDuplicateUsername) {
			return nil, status.Error(codes.AlreadyExists, "username is already taken")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate authentication token")
	}

	return &pb.RegisterResponse{Token: token}, nil
}

// Login authenticates a user by comparing the supplied password hash against the
// stored bcrypt hash. Returns a signed JWT token on success.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "authentication failed")
	}
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.PasswordHash)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate authentication token")
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *Server) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	t, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.log.Error("failed to sign jwt", zap.Error(err))
		return "", err
	}
	return t, nil
}
