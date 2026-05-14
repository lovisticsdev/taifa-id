package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-id/internal/audit"
	"taifa-id/internal/platform/clock"
	"taifa-id/internal/platform/token"
)

const (
	activeStatus    = "ACTIVE"
	tokenTypeBearer = "Bearer"
)

type PasswordVerifier interface {
	Verify(plaintext string, hash string) bool
}

type TokenManager interface {
	Issue(personID string, credentialID string, username string) (string, time.Time, string, error)
	Verify(rawToken string) (token.Claims, error)
}

type Service struct {
	pool     *pgxpool.Pool
	repo     *Repository
	verifier PasswordVerifier
	tokens   TokenManager
	clock    clock.Clock
}

func NewService(pool *pgxpool.Pool, repo *Repository, verifier PasswordVerifier, tokens TokenManager, clk clock.Clock) *Service {
	if clk == nil {
		clk = clock.NewRealClock()
	}

	return &Service{
		pool:     pool,
		repo:     repo,
		verifier: verifier,
		tokens:   tokens,
		clock:    clk,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest, correlationID string) (LoginResult, error) {
	username := normalizeUsername(req.Username)
	password := req.Password

	if username == "" || password == "" {
		return LoginResult{}, ErrValidation
	}

	record, err := s.repo.GetCredentialByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrInvalidCredential) {
			if auditErr := s.writeAuthFailed(ctx, correlationID, username, "", "", "invalid_credentials"); auditErr != nil {
				return LoginResult{}, auditErr
			}

			return LoginResult{}, ErrInvalidCredential
		}

		return LoginResult{}, err
	}

	if record.CredentialStatus != activeStatus {
		if auditErr := s.writeAuthFailed(ctx, correlationID, username, record.PersonID, record.ID, "credential_not_active"); auditErr != nil {
			return LoginResult{}, auditErr
		}

		return LoginResult{}, ErrInvalidCredential
	}

	if record.PersonStatus != activeStatus {
		if auditErr := s.writeAuthFailed(ctx, correlationID, username, record.PersonID, record.ID, "person_not_active"); auditErr != nil {
			return LoginResult{}, auditErr
		}

		return LoginResult{}, ErrInvalidCredential
	}

	if !s.verifier.Verify(password, record.PasswordHash) {
		if auditErr := s.writeAuthFailed(ctx, correlationID, username, record.PersonID, record.ID, "invalid_credentials"); auditErr != nil {
			return LoginResult{}, auditErr
		}

		return LoginResult{}, ErrInvalidCredential
	}

	accessToken, expiresAt, sessionID, err := s.tokens.Issue(record.PersonID, record.ID, record.Username)
	if err != nil {
		return LoginResult{}, fmt.Errorf("%w: %v", ErrTokenIssue, err)
	}

	event := audit.Event{
		EventType:     audit.EventAuthSucceeded,
		SubjectID:     record.PersonID,
		ResourceType:  audit.ResourceAuthSession,
		ResourceID:    sessionID,
		Action:        audit.ActionAuthenticate,
		Result:        audit.ResultSuccess,
		CorrelationID: correlationID,
		Payload: map[string]any{
			"person_id":     record.PersonID,
			"credential_id": record.ID,
			"username":      record.Username,
			"session_id":    sessionID,
			"expires_at":    expiresAt,
		},
		CreatedAt: s.clock.Now(),
	}

	if err := audit.Insert(ctx, s.pool, event); err != nil {
		return LoginResult{}, fmt.Errorf("write auth succeeded audit event: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		TokenType:    tokenTypeBearer,
		ExpiresAt:    expiresAt,
		SessionID:    sessionID,
		PersonID:     record.PersonID,
		CredentialID: record.ID,
		Username:     record.Username,
	}, nil
}

func (s *Service) Introspect(ctx context.Context, req IntrospectRequest, correlationID string) (IntrospectionResult, error) {
	rawToken := strings.TrimSpace(req.Token)
	if rawToken == "" {
		return IntrospectionResult{}, ErrValidation
	}

	claims, err := s.tokens.Verify(rawToken)
	if err != nil {
		if auditErr := s.writeIntrospectionAudit(ctx, correlationID, "", "", "", audit.ResultDenied, "invalid_token"); auditErr != nil {
			return IntrospectionResult{}, auditErr
		}

		return IntrospectionResult{Active: false}, nil
	}

	record, err := s.repo.GetCredentialByID(ctx, claims.CredentialID)
	if err != nil {
		if auditErr := s.writeIntrospectionAudit(ctx, correlationID, claims.PersonID, claims.CredentialID, claims.SessionID, audit.ResultDenied, "credential_not_found"); auditErr != nil {
			return IntrospectionResult{}, auditErr
		}

		return IntrospectionResult{Active: false}, nil
	}

	if record.CredentialStatus != activeStatus {
		if auditErr := s.writeIntrospectionAudit(ctx, correlationID, claims.PersonID, claims.CredentialID, claims.SessionID, audit.ResultDenied, "credential_not_active"); auditErr != nil {
			return IntrospectionResult{}, auditErr
		}

		return IntrospectionResult{Active: false}, nil
	}

	if record.PersonStatus != activeStatus {
		if auditErr := s.writeIntrospectionAudit(ctx, correlationID, claims.PersonID, claims.CredentialID, claims.SessionID, audit.ResultDenied, "person_not_active"); auditErr != nil {
			return IntrospectionResult{}, auditErr
		}

		return IntrospectionResult{Active: false}, nil
	}

	if err := s.writeIntrospectionAudit(ctx, correlationID, claims.PersonID, claims.CredentialID, claims.SessionID, audit.ResultAllowed, "active"); err != nil {
		return IntrospectionResult{}, err
	}

	var issuedAt *time.Time
	var expiresAt *time.Time

	if claims.IssuedAt != nil {
		value := claims.IssuedAt.Time
		issuedAt = &value
	}

	if claims.ExpiresAt != nil {
		value := claims.ExpiresAt.Time
		expiresAt = &value
	}

	return IntrospectionResult{
		Active:       true,
		PersonID:     claims.PersonID,
		CredentialID: claims.CredentialID,
		Username:     claims.Username,
		SessionID:    claims.SessionID,
		IssuedAt:     issuedAt,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *Service) writeAuthFailed(ctx context.Context, correlationID string, username string, personID string, credentialID string, reason string) error {
	event := audit.Event{
		EventType:     audit.EventAuthFailed,
		SubjectID:     personID,
		ResourceType:  audit.ResourceAuthSession,
		ResourceID:    credentialID,
		Action:        audit.ActionAuthenticate,
		Result:        audit.ResultFailure,
		CorrelationID: correlationID,
		Payload: map[string]any{
			"username":      username,
			"person_id":     personID,
			"credential_id": credentialID,
			"reason":        reason,
		},
		CreatedAt: s.clock.Now(),
	}

	if err := audit.Insert(ctx, s.pool, event); err != nil {
		return fmt.Errorf("write auth failed audit event: %w", err)
	}

	return nil
}

func (s *Service) writeIntrospectionAudit(ctx context.Context, correlationID string, personID string, credentialID string, sessionID string, result string, reason string) error {
	event := audit.Event{
		EventType:     audit.EventTokenIntrospected,
		SubjectID:     personID,
		ResourceType:  audit.ResourceAuthSession,
		ResourceID:    sessionID,
		Action:        audit.ActionIntrospectToken,
		Result:        result,
		CorrelationID: correlationID,
		Payload: map[string]any{
			"person_id":     personID,
			"credential_id": credentialID,
			"session_id":    sessionID,
			"reason":        reason,
		},
		CreatedAt: s.clock.Now(),
	}

	if err := audit.Insert(ctx, s.pool, event); err != nil {
		return fmt.Errorf("write token introspection audit event: %w", err)
	}

	return nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
