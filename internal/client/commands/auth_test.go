package commands

import (
	"context"
	"errors"
	"os"
	"testing"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/app"
	"github.com/tigranqic/gophkeeper-diploma/internal/client/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAPIClient is a test double for app.APIClient.
type mockAPIClient struct {
	registerFn func(ctx context.Context, username, passwordHash string, salt []byte) (string, error)
	getSaltFn  func(ctx context.Context, username string) ([]byte, error)
	loginFn    func(ctx context.Context, username, passwordHash string) (string, error)
	setTokenFn func(token string)
	syncFn     func(ctx context.Context, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error)
}

func (m *mockAPIClient) Register(ctx context.Context, username, passwordHash string, salt []byte) (string, error) {
	return m.registerFn(ctx, username, passwordHash, salt)
}
func (m *mockAPIClient) GetSalt(ctx context.Context, username string) ([]byte, error) {
	if m.getSaltFn != nil {
		return m.getSaltFn(ctx, username)
	}
	return []byte("16-byte-salt-xxx"), nil
}
func (m *mockAPIClient) Login(ctx context.Context, username, passwordHash string) (string, error) {
	return m.loginFn(ctx, username, passwordHash)
}
func (m *mockAPIClient) SetToken(token string) {
	if m.setTokenFn != nil {
		m.setTokenFn(token)
	}
}
func (m *mockAPIClient) Sync(ctx context.Context, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error) {
	if m.syncFn != nil {
		return m.syncFn(ctx, records, lastRevision)
	}
	return nil, 0, nil
}
func (m *mockAPIClient) Close() error { return nil }

func newTestApp(t *testing.T, client app.APIClient) (*app.App, string) {
	t.Helper()
	tmp, err := os.CreateTemp("", "gophkeeper_cmd_test_*.json")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Remove(tmp.Name()) })

	store, err := storage.Load(tmp.Name())
	require.NoError(t, err)

	// Pre-populate salt so Login can derive keys without a real TTY.
	store.Salt = []byte("16-byte-salt-xxx")

	callCount := 0
	passwords := [][]byte{[]byte("server-pass"), []byte("master-pass")}
	a := &app.App{
		Client:  client,
		Storage: store,
		ReadPassword: func(prompt string) ([]byte, error) {
			p := passwords[callCount%len(passwords)]
			callCount++
			return p, nil
		},
	}
	return a, tmp.Name()
}

func TestRegisterCmd_Success(t *testing.T) {
	client := &mockAPIClient{
		registerFn: func(_ context.Context, username, _ string, _ []byte) (string, error) {
			assert.Equal(t, "alice", username)
			return "jwt-token", nil
		},
	}
	a, _ := newTestApp(t, client)

	cmd := RegisterCmd(a)
	cmd.SetArgs([]string{"-u", "alice"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "jwt-token", a.Storage.Token)
	assert.NotEmpty(t, a.Storage.Salt)
}

func TestRegisterCmd_ServerError(t *testing.T) {
	client := &mockAPIClient{
		registerFn: func(_ context.Context, _, _ string, _ []byte) (string, error) {
			return "", errors.New("server error")
		},
	}
	a, _ := newTestApp(t, client)

	cmd := RegisterCmd(a)
	cmd.SetArgs([]string{"-u", "alice"})
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestLoginCmd_Success(t *testing.T) {
	client := &mockAPIClient{
		loginFn: func(_ context.Context, username, _ string) (string, error) {
			assert.Equal(t, "alice", username)
			return "new-token", nil
		},
	}
	a, _ := newTestApp(t, client)
	// Only one password prompt for login (master password).
	a.ReadPassword = func(prompt string) ([]byte, error) {
		return []byte("master-pass"), nil
	}

	cmd := LoginCmd(a)
	cmd.SetArgs([]string{"-u", "alice"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "new-token", a.Storage.Token)
}

func TestLoginCmd_ServerError(t *testing.T) {
	client := &mockAPIClient{
		loginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", errors.New("auth failed")
		},
	}
	a, _ := newTestApp(t, client)
	a.ReadPassword = func(prompt string) ([]byte, error) {
		return []byte("master-pass"), nil
	}

	cmd := LoginCmd(a)
	cmd.SetArgs([]string{"-u", "alice"})
	err := cmd.Execute()
	assert.Error(t, err)
}
