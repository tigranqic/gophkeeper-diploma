package gservice

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository/mocks"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestAuthService_Register_Table(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		password      string
		repoResult    *repository.User
		repoError     error
		expectedCode  codes.Code
	}{
		{
			name:         "Success",
			username:     "user1",
			password:     "pass1",
			repoResult:   &repository.User{ID: uuid.New(), Username: "user1"},
			repoError:    nil,
			expectedCode: codes.OK,
		},
		{
			name:         "Missing Username",
			username:     "",
			password:     "pass1",
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepository)
			server := NewServer(mockRepo, "secret", zap.NewNop())

			if tt.expectedCode == codes.OK || tt.name == "Internal Error" {
				mockRepo.On("CreateUser", mock.Anything, tt.username, mock.Anything).Return(tt.repoResult, tt.repoError)
			}

			req := &pb.RegisterRequest{Username: tt.username, PasswordHash: tt.password}
			res, err := server.Register(context.Background(), req)

			if tt.expectedCode == codes.OK {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Token)
			} else {
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestAuthService_Login_Table(t *testing.T) {
	password := "pass1"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		username     string
		password     string
		repoUser     *repository.User
		repoError    error
		expectedCode codes.Code
	}{
		{
			name:         "Success",
			username:     "user1",
			password:     password,
			repoUser:     &repository.User{ID: uuid.New(), Username: "user1", PasswordHash: string(hashed)},
			expectedCode: codes.OK,
		},
		{
			name:         "Invalid Password",
			username:     "user1",
			password:     "wrong",
			repoUser:     &repository.User{ID: uuid.New(), Username: "user1", PasswordHash: string(hashed)},
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "User Not Found",
			username:     "unknown",
			password:     "pass1",
			repoUser:     nil,
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockRepository)
			server := NewServer(mockRepo, "secret", zap.NewNop())

			mockRepo.On("GetUserByUsername", mock.Anything, tt.username).Return(tt.repoUser, tt.repoError)

			req := &pb.LoginRequest{Username: tt.username, PasswordHash: tt.password}
			res, err := server.Login(context.Background(), req)

			if tt.expectedCode == codes.OK {
				require.NoError(t, err)
				assert.NotEmpty(t, res.Token)
			} else {
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
