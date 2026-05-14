package person

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const postgresUniqueViolationCode = "23505"

type Queryer interface {
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

func (r *Repository) Create(ctx context.Context, db ExecerQueryer, p Person) (Person, error) {
	const query = `
		INSERT INTO persons (
			id,
			synthetic_nin,
			display_name,
			status
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			synthetic_nin,
			display_name,
			status,
			created_at,
			updated_at
	`

	var created Person
	if err := scanPerson(
		db.QueryRow(
			ctx,
			query,
			p.ID,
			p.SyntheticNIN,
			p.DisplayName,
			string(p.Status),
		),
		&created,
	); err != nil {
		if isUniqueViolation(err) {
			return Person{}, ErrDuplicateSyntheticNIN
		}

		return Person{}, fmt.Errorf("create person: %w", err)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Person, error) {
	const query = `
		SELECT
			id,
			synthetic_nin,
			display_name,
			status,
			created_at,
			updated_at
		FROM persons
		WHERE id = $1
	`

	var p Person
	if err := scanPerson(r.pool.QueryRow(ctx, query, id), &p); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Person{}, ErrNotFound
		}

		return Person{}, fmt.Errorf("get person by id: %w", err)
	}

	return p, nil
}

func (r *Repository) GetBySyntheticNIN(ctx context.Context, syntheticNIN string) (Person, error) {
	const query = `
		SELECT
			id,
			synthetic_nin,
			display_name,
			status,
			created_at,
			updated_at
		FROM persons
		WHERE synthetic_nin = $1
	`

	var p Person
	if err := scanPerson(r.pool.QueryRow(ctx, query, syntheticNIN), &p); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Person{}, ErrNotFound
		}

		return Person{}, fmt.Errorf("get person by synthetic nin: %w", err)
	}

	return p, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, db ExecerQueryer, id string, status Status) (Person, error) {
	const query = `
		UPDATE persons
		SET status = $2
		WHERE id = $1
		RETURNING
			id,
			synthetic_nin,
			display_name,
			status,
			created_at,
			updated_at
	`

	var updated Person
	if err := scanPerson(
		db.QueryRow(ctx, query, id, string(status)),
		&updated,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Person{}, ErrNotFound
		}

		return Person{}, fmt.Errorf("update person status: %w", err)
	}

	return updated, nil
}

func scanPerson(row pgx.Row, p *Person) error {
	var status string

	if err := row.Scan(
		&p.ID,
		&p.SyntheticNIN,
		&p.DisplayName,
		&status,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return err
	}

	p.Status = Status(status)
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolationCode
}
