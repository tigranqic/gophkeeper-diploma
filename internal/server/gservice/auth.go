package gservice

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/postgres"
	"github.com/tigranqic/gophkeeper-diploma/pkg/options"
)

// Server implements the AuthService and StorageService gRPC interfaces.
type Server struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedStorageServiceServer
	users     repository.UserRepository   // used by Register, Login
	records   repository.RecordRepository // used by Sync
	jwtSecret string
	log       *zap.Logger
}

// ServerOption configures a Server. Use the With* constructors below.
type ServerOption = options.Option[Server]

// WithJWTSecret sets the HMAC-SHA256 secret used to sign and verify tokens.
func WithJWTSecret(secret string) ServerOption {
	return func(s *Server) { s.jwtSecret = secret }
}

// WithLogger sets the structured logger. Defaults to zap.NewNop().
func WithLogger(log *zap.Logger) ServerOption {
	return func(s *Server) { s.log = log }
}

// NewServer creates a new Server.
// users handles auth operations; records handles storage sync.
// Options are applied after the defaults, so later options win.
//
// Defaults: jwtSecret="change-me-in-production", log=zap.NewNop().
func NewServer(users repository.UserRepository, records repository.RecordRepository, opts ...ServerOption) *Server {
	s := &Server{
		users:     users,
		records:   records,
		jwtSecret: "change-me-in-production",
		log:       zap.NewNop(),
	}
	options.Apply(s, opts)
	return s
}

// Register creates a new user account. The incoming password hash is hashed again
// with bcrypt before storage. Returns a signed JWT token on success.
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Username == "" || req.PasswordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}
	if len(req.Salt) == 0 {
		return nil, status.Error(codes.InvalidArgument, "salt is required")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to process request")
	}

	user, err := s.users.CreateUser(ctx, req.Username, string(hashed), req.Salt)
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

// GetSalt returns the scrypt salt for the given username.
func (s *Server) GetSalt(ctx context.Context, req *pb.GetSaltRequest) (*pb.GetSaltResponse, error) {
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	salt, err := s.users.GetSaltByUsername(ctx, req.Username)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to retrieve salt")
	}
	if salt == nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.GetSaltResponse{Salt: salt}, nil
}

// Login authenticates a user by comparing the supplied password hash against the
// stored bcrypt hash. Returns a signed JWT token on success.
func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.users.GetUserByUsername(ctx, req.Username)
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
