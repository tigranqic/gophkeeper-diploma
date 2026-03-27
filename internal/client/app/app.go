// Package app holds the shared application state passed to all CLI commands.
package app

import (
	"context"
	"fmt"

	"github.com/tigranqic/gophkeeper-diploma/internal/client/storage"
	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/pkg/crypto"
)

// APIClient defines the interface for communicating with the GophKeeper server.
type APIClient interface {
	// Register creates a new user account and returns a JWT token.
	Register(ctx context.Context, username, passwordHash string, salt []byte) (string, error)
	// GetSalt retrieves the scrypt salt for the given username.
	GetSalt(ctx context.Context, username string) ([]byte, error)
	// Login authenticates an existing user and returns a JWT token.
	Login(ctx context.Context, username, passwordHash string) (string, error)
	// SetToken updates the bearer token used for authenticated requests.
	SetToken(token string)
	// Sync pushes local records to the server and retrieves updates since lastRevision.
	Sync(ctx context.Context, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error)
	// Close releases the underlying network connection.
	Close() error
}

// App holds the client-side application state shared across all CLI commands.
type App struct {
	// Client communicates with the remote GophKeeper gRPC server.
	Client APIClient
	// Storage is the local encrypted record cache persisted to disk.
	Storage *storage.LocalStore
	// DEK is the Data Encryption Key derived from the user's master password.
	DEK []byte
	// ReadPassword reads a password from the user without echoing it to the terminal.
	// It is injectable so that commands can be unit-tested without a real TTY.
	ReadPassword func(prompt string) ([]byte, error)
}

// EnsureDEK checks that the DEK is initialized. If not, it prompts for the
// master password and re-derives the DEK from the locally stored salt.
// This is needed because the DEK is never persisted to disk and is lost
// between CLI invocations.
func (a *App) EnsureDEK() error {
	if len(a.DEK) > 0 {
		return nil
	}
	if len(a.Storage.Salt) == 0 {
		return fmt.Errorf("not registered: please run 'register' first")
	}
	masterPass, err := a.ReadPassword("Enter master password: ")
	if err != nil {
		return err
	}
	defer crypto.Zero(masterPass)

	dek, _, err := crypto.DeriveKeys(string(masterPass), a.Storage.Salt)
	if err != nil {
		return err
	}
	a.DEK = dek
	return nil
}
