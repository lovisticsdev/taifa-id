package organization

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

func (r *Repository) Create(ctx context.Context, db ExecerQueryer, org Organization) (Organization, error) {
	const query = `
		INSERT INTO organizations (
			id,
			name,
			primary_type,
			status
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			name,
			primary_type,
			status,
			created_at,
			updated_at
	`

	var created Organization
	if err := scanOrganization(
		db.QueryRow(
			ctx,
			query,
			org.ID,
			org.Name,
			string(org.PrimaryType),
			string(org.Status),
		),
		&created,
	); err != nil {
		return Organization{}, fmt.Errorf("create organization: %w", err)
	}

	return created, nil
}

func (r *Repository) List(ctx context.Context) ([]Organization, error) {
	const query = `
		SELECT
			id,
			name,
			primary_type,
			status,
			created_at,
			updated_at
		FROM organizations
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	orgs := make([]Organization, 0)
	for rows.Next() {
		var org Organization
		if err := scanOrganization(rows, &org); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}

		orgs = append(orgs, org)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizations: %w", err)
	}

	return orgs, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Organization, error) {
	const query = `
		SELECT
			id,
			name,
			primary_type,
			status,
			created_at,
			updated_at
		FROM organizations
		WHERE id = $1
	`

	var org Organization
	if err := scanOrganization(r.pool.QueryRow(ctx, query, id), &org); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}

		return Organization{}, fmt.Errorf("get organization by id: %w", err)
	}

	return org, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, db ExecerQueryer, id string, status Status) (Organization, error) {
	const query = `
		UPDATE organizations
		SET status = $2
		WHERE id = $1
		RETURNING
			id,
			name,
			primary_type,
			status,
			created_at,
			updated_at
	`

	var updated Organization
	if err := scanOrganization(db.QueryRow(ctx, query, id, string(status)), &updated); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}

		return Organization{}, fmt.Errorf("update organization status: %w", err)
	}

	return updated, nil
}

func (r *Repository) AddCapability(ctx context.Context, db ExecerQueryer, capability OrganizationCapability) (OrganizationCapability, error) {
	const query = `
		INSERT INTO organization_capabilities (
			id,
			organization_id,
			capability
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			organization_id,
			capability,
			created_at
	`

	var created OrganizationCapability
	if err := scanCapability(
		db.QueryRow(
			ctx,
			query,
			capability.ID,
			capability.OrganizationID,
			string(capability.Capability),
		),
		&created,
	); err != nil {
		if isUniqueViolation(err) {
			return OrganizationCapability{}, ErrDuplicateCapability
		}

		if isForeignKeyViolation(err) {
			return OrganizationCapability{}, ErrNotFound
		}

		return OrganizationCapability{}, fmt.Errorf("add organization capability: %w", err)
	}

	return created, nil
}

func (r *Repository) ListCapabilities(ctx context.Context, organizationID string) ([]OrganizationCapability, error) {
	const query = `
		SELECT
			id,
			organization_id,
			capability,
			created_at
		FROM organization_capabilities
		WHERE organization_id = $1
		ORDER BY capability ASC
	`

	rows, err := r.pool.Query(ctx, query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization capabilities: %w", err)
	}
	defer rows.Close()

	capabilities := make([]OrganizationCapability, 0)
	for rows.Next() {
		var capability OrganizationCapability
		if err := scanCapability(rows, &capability); err != nil {
			return nil, fmt.Errorf("scan organization capability: %w", err)
		}

		capabilities = append(capabilities, capability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organization capabilities: %w", err)
	}

	return capabilities, nil
}

func (r *Repository) RemoveCapability(ctx context.Context, db ExecerQueryer, organizationID string, capability Capability) (OrganizationCapability, error) {
	const query = `
		DELETE FROM organization_capabilities
		WHERE organization_id = $1
		  AND capability = $2
		RETURNING
			id,
			organization_id,
			capability,
			created_at
	`

	var removed OrganizationCapability
	if err := scanCapability(
		db.QueryRow(ctx, query, organizationID, string(capability)),
		&removed,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrganizationCapability{}, ErrCapabilityNotFound
		}

		return OrganizationCapability{}, fmt.Errorf("remove organization capability: %w", err)
	}

	return removed, nil
}

func scanOrganization(row pgx.Row, org *Organization) error {
	var primaryType string
	var status string

	if err := row.Scan(
		&org.ID,
		&org.Name,
		&primaryType,
		&status,
		&org.CreatedAt,
		&org.UpdatedAt,
	); err != nil {
		return err
	}

	org.PrimaryType = PrimaryType(primaryType)
	org.Status = Status(status)

	return nil
}

func scanCapability(row pgx.Row, capability *OrganizationCapability) error {
	var capabilityValue string

	if err := row.Scan(
		&capability.ID,
		&capability.OrganizationID,
		&capabilityValue,
		&capability.CreatedAt,
	); err != nil {
		return err
	}

	capability.Capability = Capability(capabilityValue)

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
