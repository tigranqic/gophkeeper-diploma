package api

import (
	"context"
	"io"
	"time"

	pb "github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const defaultTimeout = 30 * time.Second

// Client is a wrapper for gRPC clients.
type Client struct {
	authClient    pb.AuthServiceClient
	storageClient pb.StorageServiceClient
	token         string
}

// NewClient creates a new gRPC client.
func NewClient(addr string, token string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		authClient:    pb.NewAuthServiceClient(conn),
		storageClient: pb.NewStorageServiceClient(conn),
		token:         token,
	}, nil
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
func (c *Client) Register(ctx context.Context, username, passwordHash string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	res, err := c.authClient.Register(ctx, &pb.RegisterRequest{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return "", err
	}
	return res.Token, nil
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
	var currentRevision int64 = lastRevision

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
