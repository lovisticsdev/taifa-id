package organization

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/ids"
	"taifa-id/internal/platform/postgres"
)

type Service struct {
	pool  *pgxpool.Pool
	repo  *Repository
	clock clock.Clock
}

func NewService(pool *pgxpool.Pool, repo *Repository, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.NewRealClock()
	}

	return &Service{
		pool:  pool,
		repo:  repo,
		clock: clk,
	}
}

func (s *Service) Create(ctx context.Context, req CreateOrganizationRequest, correlationID string) (Organization, error) {
	name := strings.TrimSpace(req.Name)
	primaryType := PrimaryType(strings.TrimSpace(req.PrimaryType))

	if name == "" || !IsValidPrimaryType(primaryType) {
		return Organization{}, ErrValidation
	}

	org := Organization{
		ID:          ids.NewOrganizationID(),
		Name:        name,
		PrimaryType: primaryType,
		Status:      StatusActive,
	}

	var created Organization

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		created, err = s.repo.Create(ctx, tx, org)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventOrganizationCreated,
			SubjectID:     created.ID,
			ResourceType:  audit.ResourceOrganization,
			ResourceID:    created.ID,
			Action:        audit.ActionCreate,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"name":         created.Name,
				"primary_type": string(created.PrimaryType),
				"status":       string(created.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write organization created audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return Organization{}, err
	}

	return created, nil
}

func (s *Service) List(ctx context.Context) ([]Organization, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (Organization, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Organization{}, ErrValidation
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdateOrganizationStatusRequest, correlationID string) (Organization, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Organization{}, ErrValidation
	}

	status := Status(strings.TrimSpace(req.Status))
	if !IsValidStatus(status) {
		return Organization{}, ErrValidation
	}

	var updated Organization

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		updated, err = s.repo.UpdateStatus(ctx, tx, id, status)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventOrganizationStatusChanged,
			SubjectID:     updated.ID,
			ResourceType:  audit.ResourceOrganization,
			ResourceID:    updated.ID,
			Action:        audit.ActionUpdateStatus,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"status": string(updated.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write organization status changed audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return Organization{}, err
	}

	return updated, nil
}

func (s *Service) AddCapability(ctx context.Context, organizationID string, req AddOrganizationCapabilityRequest, correlationID string) (OrganizationCapability, error) {
	organizationID = strings.TrimSpace(organizationID)
	capabilityValue := Capability(strings.TrimSpace(req.Capability))

	if organizationID == "" || !IsValidCapability(capabilityValue) {
		return OrganizationCapability{}, ErrValidation
	}

	capability := OrganizationCapability{
		ID:             ids.NewOrganizationCapabilityID(),
		OrganizationID: organizationID,
		Capability:     capabilityValue,
	}

	var created OrganizationCapability

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		created, err = s.repo.AddCapability(ctx, tx, capability)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventOrganizationCapabilityAdded,
			SubjectID:     organizationID,
			ResourceType:  audit.ResourceOrganizationCapability,
			ResourceID:    created.ID,
			Action:        audit.ActionAddCapability,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"organization_id": organizationID,
				"capability":      string(created.Capability),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write organization capability added audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateCapability) {
			return OrganizationCapability{}, ErrDuplicateCapability
		}

		return OrganizationCapability{}, err
	}

	return created, nil
}

func (s *Service) ListCapabilities(ctx context.Context, organizationID string) ([]OrganizationCapability, error) {
	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return nil, ErrValidation
	}

	if _, err := s.repo.GetByID(ctx, organizationID); err != nil {
		return nil, err
	}

	return s.repo.ListCapabilities(ctx, organizationID)
}

func (s *Service) RemoveCapability(ctx context.Context, organizationID string, capability string, correlationID string) (OrganizationCapability, error) {
	organizationID = strings.TrimSpace(organizationID)
	capabilityValue := Capability(strings.TrimSpace(capability))

	if organizationID == "" || !IsValidCapability(capabilityValue) {
		return OrganizationCapability{}, ErrValidation
	}

	var removed OrganizationCapability

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		removed, err = s.repo.RemoveCapability(ctx, tx, organizationID, capabilityValue)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventOrganizationCapabilityRemoved,
			SubjectID:     organizationID,
			ResourceType:  audit.ResourceOrganizationCapability,
			ResourceID:    removed.ID,
			Action:        audit.ActionRemoveCapability,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"organization_id": organizationID,
				"capability":      string(removed.Capability),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write organization capability removed audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return OrganizationCapability{}, err
	}

	return removed, nil
}
