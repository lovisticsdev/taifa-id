package actorcontext

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

func (r *Repository) GetCredentialByID(ctx context.Context, credentialID string) (CredentialRecord, error) {
	const query = `
		SELECT
			c.id,
			c.person_id,
			c.username,
			c.status,
			p.status
		FROM credentials c
		JOIN persons p ON p.id = c.person_id
		WHERE c.id = $1
	`

	var record CredentialRecord
	if err := r.pool.QueryRow(ctx, query, credentialID).Scan(
		&record.ID,
		&record.PersonID,
		&record.Username,
		&record.CredentialStatus,
		&record.PersonStatus,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CredentialRecord{}, ErrInvalidToken
		}

		return CredentialRecord{}, fmt.Errorf("get credential by id: %w", err)
	}

	return record, nil
}

func (r *Repository) OrganizationStatus(ctx context.Context, organizationID string) (string, error) {
	const query = `
		SELECT status
		FROM organizations
		WHERE id = $1
	`

	var status string
	if err := r.pool.QueryRow(ctx, query, organizationID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrOrganizationNotFound
		}

		return "", fmt.Errorf("get organization status: %w", err)
	}

	return status, nil
}

func (r *Repository) ActiveMemberships(ctx context.Context, personID string, organizationID string) ([]MembershipContext, error) {
	const query = `
		SELECT
			id,
			membership_type
		FROM organization_memberships
		WHERE person_id = $1
		  AND organization_id = $2
		  AND status = 'ACTIVE'
		  AND (ends_at IS NULL OR ends_at > now())
		ORDER BY created_at ASC, id ASC
	`

	rows, err := r.pool.Query(ctx, query, personID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list active memberships: %w", err)
	}
	defer rows.Close()

	memberships := make([]MembershipContext, 0)
	for rows.Next() {
		var membership MembershipContext
		if err := rows.Scan(
			&membership.ID,
			&membership.MembershipType,
		); err != nil {
			return nil, fmt.Errorf("scan active membership: %w", err)
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active memberships: %w", err)
	}

	if len(memberships) == 0 {
		return nil, ErrNoActiveMembership
	}

	return memberships, nil
}

func (r *Repository) RolesForActiveMemberships(ctx context.Context, personID string, organizationID string) ([]string, error) {
	const query = `
		SELECT DISTINCT mr.role
		FROM membership_roles mr
		JOIN organization_memberships om ON om.id = mr.membership_id
		WHERE om.person_id = $1
		  AND om.organization_id = $2
		  AND om.status = 'ACTIVE'
		  AND (om.ends_at IS NULL OR om.ends_at > now())
		ORDER BY mr.role ASC
	`

	rows, err := r.pool.Query(ctx, query, personID, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list roles for active memberships: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan membership role: %w", err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate membership roles: %w", err)
	}

	return roles, nil
}
