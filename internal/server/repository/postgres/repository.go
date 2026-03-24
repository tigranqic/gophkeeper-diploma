// Package postgres provides a PostgreSQL-backed implementation of the repository.Repository interface.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tigranqic/gophkeeper-diploma/internal/proto/gophkeeperv1"
	"github.com/tigranqic/gophkeeper-diploma/internal/server/repository"
	"go.uber.org/zap"
)

// ErrDuplicateUsername is returned by CreateUser when the username is already taken.
var ErrDuplicateUsername = errors.New("username already exists")

// PostgresRepository is a repository.Repository backed by a PostgreSQL connection pool.
type PostgresRepository struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

// NewPostgresRepository opens a pgxpool connection to the given DSN and pings the server.
func NewPostgresRepository(ctx context.Context, dsn string, log *zap.Logger) (*PostgresRepository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresRepository{pool: pool, log: log}, nil
}

// CreateUser inserts a new user row and returns the created User with its generated UUID.
// Returns ErrDuplicateUsername if the username is already taken.
func (r *PostgresRepository) CreateUser(ctx context.Context, username, passwordHash string) (*repository.User, error) {
	user := &repository.User{
		Username:     username,
		PasswordHash: passwordHash,
	}

	query := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id`
	err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateUsername
		}
		r.log.Error("failed to create user", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("internal database error")
	}

	return user, nil
}

// GetUserByUsername fetches a user by their unique username.
// Returns (nil, nil) when no matching user exists.
func (r *PostgresRepository) GetUserByUsername(ctx context.Context, username string) (*repository.User, error) {
	user := &repository.User{Username: username}
	query := `SELECT id, password_hash FROM users WHERE username = $1`
	err := r.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.PasswordHash)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.log.Error("failed to get user", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("internal database error")
	}

	return user, nil
}

// SyncRecords upserts the supplied records (applying last-write-wins on updated_at),
// then queries and returns all records for userID with revision > lastRevision.
// The second return value is the highest revision seen after the upsert.
func (r *PostgresRepository) SyncRecords(ctx context.Context, userID uuid.UUID, records []*gophkeeperv1.EncryptedRecord, lastRevision int64) ([]*gophkeeperv1.EncryptedRecord, int64, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback(ctx)

	// Upsert incoming records
	for _, rec := range records {
		query := `
			INSERT INTO records (id, user_id, type, blob, updated_at, is_deleted)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET type = EXCLUDED.type,
			    blob = EXCLUDED.blob,
			    updated_at = EXCLUDED.updated_at,
			    is_deleted = EXCLUDED.is_deleted,
			    revision = DEFAULT -- This triggers BIGSERIAL increment
			WHERE EXCLUDED.updated_at > records.updated_at
		`
		_, err := tx.Exec(ctx, query,
			rec.Id,
			userID,
			rec.Type,
			rec.Blob,
			rec.UpdatedAt,
			rec.IsDeleted,
		)
		if err != nil {
			r.log.Error("failed to upsert record", zap.Error(err), zap.String("user_id", userID.String()))
			return nil, 0, fmt.Errorf("sync failed")
		}
	}

	// Get server-side updates
	query := `
		SELECT id, type, blob, revision, updated_at, is_deleted
		FROM records
		WHERE user_id = $1 AND revision > $2
		ORDER BY revision ASC
	`
	rows, err := tx.Query(ctx, query, userID, lastRevision)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var updates []*gophkeeperv1.EncryptedRecord
	var maxRevision int64 = lastRevision

	for rows.Next() {
		var rec gophkeeperv1.EncryptedRecord
		err := rows.Scan(
			&rec.Id,
			&rec.Type,
			&rec.Blob,
			&rec.Revision,
			&rec.UpdatedAt,
			&rec.IsDeleted,
		)
		if err != nil {
			return nil, 0, err
		}
		updates = append(updates, &rec)
		if rec.Revision > maxRevision {
			maxRevision = rec.Revision
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, err
	}

	return updates, maxRevision, nil
}

// Close releases all connections in the pool.
func (r *PostgresRepository) Close() {
	r.pool.Close()
}
