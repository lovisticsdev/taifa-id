package membership

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *Repository) OrganizationStatus(ctx context.Context, db Queryer, organizationID string) (string, error) {
	const query = `
		SELECT status
		FROM organizations
		WHERE id = $1
	`

	var status string
	if err := db.QueryRow(ctx, query, organizationID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrReferenceNotFound
		}

		return "", fmt.Errorf("get organization status: %w", err)
	}

	return status, nil
}

func (r *Repository) Create(ctx context.Context, db ExecerQueryer, membership Membership) (Membership, error) {
	const query = `
		INSERT INTO organization_memberships (
			id,
			person_id,
			organization_id,
			membership_type,
			status
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
	`

	var created Membership
	if err := scanMembership(
		db.QueryRow(
			ctx,
			query,
			membership.ID,
			membership.PersonID,
			membership.OrganizationID,
			string(membership.MembershipType),
			string(membership.Status),
		),
		&created,
	); err != nil {
		if isUniqueViolation(err) {
			return Membership{}, ErrDuplicateActiveMembership
		}

		if isForeignKeyViolation(err) {
			return Membership{}, ErrReferenceNotFound
		}

		return Membership{}, fmt.Errorf("create membership: %w", err)
	}

	return created, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Membership, error) {
	const query = `
		SELECT
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM organization_memberships
		WHERE id = $1
	`

	var membership Membership
	if err := scanMembership(r.pool.QueryRow(ctx, query, id), &membership); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrNotFound
		}

		return Membership{}, fmt.Errorf("get membership by id: %w", err)
	}

	return membership, nil
}

func (r *Repository) ListByPerson(ctx context.Context, personID string) ([]Membership, error) {
	const query = `
		SELECT
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM organization_memberships
		WHERE person_id = $1
		ORDER BY created_at ASC, id ASC
	`

	return r.list(ctx, query, personID)
}

func (r *Repository) ListByOrganization(ctx context.Context, organizationID string) ([]Membership, error) {
	const query = `
		SELECT
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
		FROM organization_memberships
		WHERE organization_id = $1
		ORDER BY created_at ASC, id ASC
	`

	return r.list(ctx, query, organizationID)
}

func (r *Repository) UpdateStatus(ctx context.Context, db ExecerQueryer, id string, status Status) (Membership, error) {
	const query = `
		UPDATE organization_memberships
		SET
			status = $2,
			ends_at = CASE
				WHEN $2 = 'ENDED' THEN COALESCE(ends_at, now())
				ELSE ends_at
			END
		WHERE id = $1
		RETURNING
			id,
			person_id,
			organization_id,
			membership_type,
			status,
			starts_at,
			ends_at,
			created_at,
			updated_at
	`

	var updated Membership
	if err := scanMembership(db.QueryRow(ctx, query, id, string(status)), &updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Membership{}, ErrNotFound
		}

		return Membership{}, fmt.Errorf("update membership status: %w", err)
	}

	return updated, nil
}

func (r *Repository) AddRole(ctx context.Context, db ExecerQueryer, role MembershipRole) (MembershipRole, error) {
	const query = `
		INSERT INTO membership_roles (
			id,
			membership_id,
			role
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			membership_id,
			role,
			created_at
	`

	var created MembershipRole
	if err := scanMembershipRole(
		db.QueryRow(ctx, query, role.ID, role.MembershipID, string(role.Role)),
		&created,
	); err != nil {
		if isUniqueViolation(err) {
			return MembershipRole{}, ErrDuplicateRole
		}

		if isForeignKeyViolation(err) {
			return MembershipRole{}, ErrNotFound
		}

		return MembershipRole{}, fmt.Errorf("add membership role: %w", err)
	}

	return created, nil
}

func (r *Repository) ListRoles(ctx context.Context, membershipID string) ([]MembershipRole, error) {
	const query = `
		SELECT
			id,
			membership_id,
			role,
			created_at
		FROM membership_roles
		WHERE membership_id = $1
		ORDER BY role ASC
	`

	rows, err := r.pool.Query(ctx, query, membershipID)
	if err != nil {
		return nil, fmt.Errorf("list membership roles: %w", err)
	}
	defer rows.Close()

	roles := make([]MembershipRole, 0)
	for rows.Next() {
		var role MembershipRole
		if err := scanMembershipRole(rows, &role); err != nil {
			return nil, fmt.Errorf("scan membership role: %w", err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate membership roles: %w", err)
	}

	return roles, nil
}

func (r *Repository) RemoveRole(ctx context.Context, db ExecerQueryer, membershipID string, role Role) (MembershipRole, error) {
	const query = `
		DELETE FROM membership_roles
		WHERE membership_id = $1
		  AND role = $2
		RETURNING
			id,
			membership_id,
			role,
			created_at
	`

	var removed MembershipRole
	if err := scanMembershipRole(
		db.QueryRow(ctx, query, membershipID, string(role)),
		&removed,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MembershipRole{}, ErrRoleNotFound
		}

		return MembershipRole{}, fmt.Errorf("remove membership role: %w", err)
	}

	return removed, nil
}

func (r *Repository) list(ctx context.Context, query string, arg any) ([]Membership, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	defer rows.Close()

	memberships := make([]Membership, 0)
	for rows.Next() {
		var membership Membership
		if err := scanMembership(rows, &membership); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate memberships: %w", err)
	}

	return memberships, nil
}

func scanMembership(row pgx.Row, membership *Membership) error {
	var membershipType string
	var status string
	var endsAt pgtype.Timestamptz

	if err := row.Scan(
		&membership.ID,
		&membership.PersonID,
		&membership.OrganizationID,
		&membershipType,
		&status,
		&membership.StartsAt,
		&endsAt,
		&membership.CreatedAt,
		&membership.UpdatedAt,
	); err != nil {
		return err
	}

	membership.MembershipType = Type(membershipType)
	membership.Status = Status(status)
	membership.EndsAt = nil

	if endsAt.Valid {
		value := endsAt.Time
		membership.EndsAt = &value
	}

	return nil
}

func scanMembershipRole(row pgx.Row, role *MembershipRole) error {
	var roleValue string

	if err := row.Scan(
		&role.ID,
		&role.MembershipID,
		&roleValue,
		&role.CreatedAt,
	); err != nil {
		return err
	}

	role.Role = Role(roleValue)

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
