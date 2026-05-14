package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetCredentialByUsername(ctx context.Context, username string) (CredentialRecord, error) {
	const query = `
		SELECT
			c.id,
			c.person_id,
			c.username,
			c.password_hash,
			c.status,
			p.status
		FROM credentials c
		JOIN persons p ON p.id = c.person_id
		WHERE c.username = $1
	`

	return r.getCredential(ctx, query, username)
}

func (r *Repository) GetCredentialByID(ctx context.Context, credentialID string) (CredentialRecord, error) {
	const query = `
		SELECT
			c.id,
			c.person_id,
			c.username,
			c.password_hash,
			c.status,
			p.status
		FROM credentials c
		JOIN persons p ON p.id = c.person_id
		WHERE c.id = $1
	`

	return r.getCredential(ctx, query, credentialID)
}

func (r *Repository) getCredential(ctx context.Context, query string, arg any) (CredentialRecord, error) {
	var record CredentialRecord

	if err := r.pool.QueryRow(ctx, query, arg).Scan(
		&record.ID,
		&record.PersonID,
		&record.Username,
		&record.PasswordHash,
		&record.CredentialStatus,
		&record.PersonStatus,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CredentialRecord{}, ErrInvalidCredential
		}

		return CredentialRecord{}, fmt.Errorf("get credential: %w", err)
	}

	return record, nil
}
