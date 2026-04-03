package app

import (
	"context"
	"testing"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/stretchr/testify/assert"
)

// stubClient implements APIClient for testing.
type stubClient struct{}

func (s *stubClient) Register(_ context.Context, _, _ string, _ []byte) (string, error) {
	return "tok", nil
}
func (s *stubClient) GetSalt(_ context.Context, _ string) ([]byte, error) {
	return []byte("16-byte-salt-xxx"), nil
}
func (s *stubClient) Login(_ context.Context, _, _ string) (string, error) { return "tok", nil }
func (s *stubClient) SetToken(_ string)                                     {}
func (s *stubClient) Sync(_ context.Context, _ []*pb.EncryptedRecord, _ int64) ([]*pb.EncryptedRecord, int64, error) {
	return nil, 0, nil
}
func (s *stubClient) Close() error { return nil }

func TestApp_DEKField(t *testing.T) {
	a := &App{
		Client: &stubClient{},
		DEK:    []byte("secret"),
	}
	assert.Equal(t, []byte("secret"), a.DEK)
}

func TestApp_ReadPasswordField(t *testing.T) {
	called := false
	a := &App{
		Client: &stubClient{},
		ReadPassword: func(prompt string) ([]byte, error) {
			called = true
			return []byte("pass"), nil
		},
	}
	p, err := a.ReadPassword("Enter: ")
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []byte("pass"), p)
}
