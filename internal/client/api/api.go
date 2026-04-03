package api

import (
	"context"
	"io"
	"time"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/pkg/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const defaultTimeout = 30 * time.Second

// Client is a wrapper for gRPC clients.
type Client struct {
	conn          *grpc.ClientConn
	authClient    pb.AuthServiceClient
	storageClient pb.StorageServiceClient
	token         string
	timeout       time.Duration
}

// ClientOption configures a Client. Use the With* constructors below.
type ClientOption = options.Option[Client]

// WithToken sets the Bearer JWT token attached to every authenticated request.
func WithToken(token string) ClientOption {
	return func(c *Client) { c.token = token }
}

// WithTimeout overrides the per-call deadline. Defaults to 30 s.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.timeout = d }
}

// NewClient dials addr and returns a configured Client.
// Options are applied after the defaults, so later options win.
func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := &Client{
		conn:          conn,
		authClient:    pb.NewAuthServiceClient(conn),
		storageClient: pb.NewStorageServiceClient(conn),
		timeout:       defaultTimeout,
	}
	options.Apply(c, opts)
	return c, nil
}

// Close releases the underlying gRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// SetToken sets the JWT token for authentication.
func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) getCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	newCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	return metadata.AppendToOutgoingContext(newCtx, "authorization", "Bearer "+c.token), cancel
}

// Register registers a new user and returns a JWT token.
func (c *Client) Register(ctx context.Context, username, passwordHash string, salt []byte) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := c.authClient.Register(ctx, &pb.RegisterRequest{
		Username:     username,
		PasswordHash: passwordHash,
		Salt:         salt,
	})
	if err != nil {
		return "", err
	}
	return res.Token, nil
}

// GetSalt retrieves the scrypt salt for the given username from the server.
func (c *Client) GetSalt(ctx context.Context, username string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := c.authClient.GetSalt(ctx, &pb.GetSaltRequest{Username: username})
	if err != nil {
		return nil, err
	}
	return res.Salt, nil
}

// Login authenticates a user and returns a JWT token.
func (c *Client) Login(ctx context.Context, username, passwordHash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := c.authClient.Login(ctx, &pb.LoginRequest{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return "", err
	}
	return res.Token, nil
}

// Sync synchronizes records with the server using streaming.
func (c *Client) Sync(ctx context.Context, records []*pb.EncryptedRecord, lastRevision int64) ([]*pb.EncryptedRecord, int64, error) {
	ctx, cancel := c.getCtx(ctx)
	defer cancel()

	stream, err := c.storageClient.Sync(ctx, &pb.SyncRequest{
		Records:      records,
		LastRevision: lastRevision,
	})
	if err != nil {
		return nil, 0, err
	}

	var allUpdates []*pb.EncryptedRecord
	currentRevision := lastRevision

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		allUpdates = append(allUpdates, res.Records...)
		if res.CurrentRevision > currentRevision {
			currentRevision = res.CurrentRevision
		}
	}

	return allUpdates, currentRevision, nil
}
