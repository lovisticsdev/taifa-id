package credential

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/ids"
	"taifa-id/internal/platform/postgres"
)

const (
	activeStatus      = "ACTIVE"
	minPasswordLength = 8
	maxPasswordLength = 256
)

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
}

type Service struct {
	pool   *pgxpool.Pool
	repo   *Repository
	hasher PasswordHasher
	clock  clock.Clock
}

func NewService(pool *pgxpool.Pool, repo *Repository, hasher PasswordHasher, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.NewRealClock()
	}

	return &Service{
		pool:   pool,
		repo:   repo,
		hasher: hasher,
		clock:  clk,
	}
}

func (s *Service) Create(ctx context.Context, req CreateCredentialRequest, correlationID string) (Credential, error) {
	personID := strings.TrimSpace(req.PersonID)
	username := normalizeUsername(req.Username)

	if personID == "" || username == "" || !isValidPassword(req.Password) {
		return Credential{}, ErrValidation
	}

	passwordHash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return Credential{}, fmt.Errorf("%w: %v", ErrHashPassword, err)
	}

	credential := Credential{
		ID:           ids.NewCredentialID(),
		PersonID:     personID,
		Username:     username,
		PasswordHash: passwordHash,
		Status:       StatusActive,
	}

	var created Credential

	err = postgres.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		personStatus, err := s.repo.PersonStatus(ctx, tx, personID)
		if err != nil {
			return err
		}

		if personStatus != activeStatus {
			return ErrPersonNotActive
		}

		created, err = s.repo.Create(ctx, tx, credential)
		if err != nil {
			return err
		}

		event := audit.Event{
			EventType:     audit.EventCredentialCreated,
			SubjectID:     created.PersonID,
			ResourceType:  audit.ResourceCredential,
			ResourceID:    created.ID,
			Action:        audit.ActionCreate,
			Result:        audit.ResultSuccess,
			CorrelationID: correlationID,
			Payload: map[string]any{
				"person_id": created.PersonID,
				"username":  created.Username,
				"status":    string(created.Status),
			},
			CreatedAt: s.clock.Now(),
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return fmt.Errorf("write credential created audit event: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrDuplicateUsername) {
			return Credential{}, ErrDuplicateUsername
		}

		return Credential{}, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (Credential, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Credential{}, ErrValidation
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByPerson(ctx context.Context, personID string) ([]Credential, error) {
	personID = strings.TrimSpace(personID)
	if personID == "" {
		return nil, ErrValidation
	}

	return s.repo.ListByPerson(ctx, personID)
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func isValidPassword(password string) bool {
	length := utf8.RuneCountInString(password)
	return length >= minPasswordLength && length <= maxPasswordLength
}
