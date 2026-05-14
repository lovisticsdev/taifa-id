package credential

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	postgresUniqueViolationCode     = "23505"
	postgresForeignKeyViolationCode = "23503"
)

type Queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type ExecerQueryer interface {
	Queryer
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) PersonStatus(ctx context.Context, db Queryer, personID string) (string, error) {
	const query = `
		SELECT status
		FROM persons
		WHERE id = $1
	`

	var status string
	if err := db.QueryRow(ctx, query, personID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrReferenceNotFound
		}

		return "", fmt.Errorf("get person status: %w", err)
	}

	return status, nil
}

func (r *Repository) Create(ctx context.Context, db ExecerQueryer, credential Credential) (Credential, error) {
	const query = `
		INSERT INTO credentials (
			id,
			person_id,
			username,
			password_hash,
			status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			person_id,
			username,
			password_hash,
			status,
			created_at,
			updated_at
	`

	var created Credential
	if err := scanCredential(
		db.QueryRow(
			ctx,
			query,
			credential.ID,
			credential.PersonID,
			credential.Username,
			credential.PasswordHash,
			string(credential.Status),
		),
		&created,
	); err != nil {
		if isUniqueViolation(err) {
			return Credential{}, ErrDuplicateUsername
		}

		if isForeignKeyViolation(err) {
			return Credential{}, ErrReferenceNotFound
		}

		return Credential{}, fmt.Errorf("create credential: %w", err)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Credential, error) {
	const query = `
		SELECT
			id,
			person_id,
			username,
			password_hash,
			status,
			created_at,
			updated_at
		FROM credentials
		WHERE id = $1
	`

	var credential Credential
	if err := scanCredential(r.pool.QueryRow(ctx, query, id), &credential); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Credential{}, ErrNotFound
		}

		return Credential{}, fmt.Errorf("get credential by id: %w", err)
	}

	return credential, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (Credential, error) {
	const query = `
		SELECT
			id,
			person_id,
			username,
			password_hash,
			status,
			created_at,
			updated_at
		FROM credentials
		WHERE username = $1
	`

	var credential Credential
	if err := scanCredential(r.pool.QueryRow(ctx, query, username), &credential); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Credential{}, ErrNotFound
		}

		return Credential{}, fmt.Errorf("get credential by username: %w", err)
	}

	return credential, nil
}

func (r *Repository) ListByPerson(ctx context.Context, personID string) ([]Credential, error) {
	const query = `
		SELECT
			id,
			person_id,
			username,
			password_hash,
			status,
			created_at,
			updated_at
		FROM credentials
		WHERE person_id = $1
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, personID)
	if err != nil {
		return nil, fmt.Errorf("list credentials by person: %w", err)
	}
	defer rows.Close()

	credentials := make([]Credential, 0)
	for rows.Next() {
		var credential Credential
		if err := scanCredential(rows, &credential); err != nil {
			return nil, fmt.Errorf("scan credential: %w", err)
		}

		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}

	return credentials, nil
}

func scanCredential(row pgx.Row, credential *Credential) error {
	var status string

	if err := row.Scan(
		&credential.ID,
		&credential.PersonID,
		&credential.Username,
		&credential.PasswordHash,
		&status,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	); err != nil {
		return err
	}

	credential.Status = Status(status)

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolationCode
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresForeignKeyViolationCode
}
