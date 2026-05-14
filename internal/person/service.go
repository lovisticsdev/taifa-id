package person

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

func (s *Service) Create(ctx context.Context, req CreatePersonRequest, correlationID string) (Person, error) {
	syntheticNIN := strings.TrimSpace(req.SyntheticNIN)
	displayName := strings.TrimSpace(req.DisplayName)

	if syntheticNIN == "" || displayName == "" {
		return Person{}, ErrValidation
	}

	p := Person{
		ID:           ids.NewPersonID(),
		SyntheticNIN: syntheticNIN,
		DisplayName:  displayName,
		Status:       StatusActive,
	}

	var created Person

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		created, err = s.repo.Create(ctx, tx, p)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventPersonCreated,
			SubjectID:     created.ID,
			ResourceType:  audit.ResourcePerson,
			ResourceID:    created.ID,
			Action:        audit.ActionCreate,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"synthetic_nin": created.SyntheticNIN,
				"display_name":  created.DisplayName,
				"status":        string(created.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write person created audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateSyntheticNIN) {
			return Person{}, ErrDuplicateSyntheticNIN
		}

		return Person{}, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Person, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Person{}, ErrValidation
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetBySyntheticNIN(ctx context.Context, syntheticNIN string) (Person, error) {
	syntheticNIN = strings.TrimSpace(syntheticNIN)
	if syntheticNIN == "" {
		return Person{}, ErrValidation
	}

	return s.repo.GetBySyntheticNIN(ctx, syntheticNIN)
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req UpdatePersonStatusRequest, correlationID string) (Person, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Person{}, ErrValidation
	}

	status := Status(strings.TrimSpace(req.Status))
	if !IsValidStatus(status) {
		return Person{}, ErrValidation
	}

	var updated Person

	err := postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error

		updated, err = s.repo.UpdateStatus(ctx, tx, id, status)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventPersonStatusChanged,
			SubjectID:     updated.ID,
			ResourceType:  audit.ResourcePerson,
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
			return fmt.Errorf("write person status changed audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		return Person{}, err
	}

	return updated, nil
}
