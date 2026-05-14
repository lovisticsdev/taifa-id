package token

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"taifa-id/internal/platform/ids"
)

var (
	ErrInvalidConfig           = errors.New("invalid jwt config")
	ErrInvalidToken            = errors.New("invalid token")
	ErrUnexpectedSigningMethod = errors.New("unexpected jwt signing method")
)

type Config struct {
	Secret   string
	Issuer   string
	Audience string
	TTL      time.Duration
}

type Claims struct {
	PersonID     string `json:"person_id"`
	CredentialID string `json:"credential_id"`
	Username     string `json:"username"`
	SessionID    string `json:"session_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWTManager(cfg Config) (*JWTManager, error) {
	secret := strings.TrimSpace(cfg.Secret)
	issuer := strings.TrimSpace(cfg.Issuer)
	audience := strings.TrimSpace(cfg.Audience)

	if len(secret) < 32 || issuer == "" || audience == "" || cfg.TTL <= 0 {
		return nil, ErrInvalidConfig
	}

	return &JWTManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      cfg.TTL,
	}, nil
}

func (m *JWTManager) Issue(personID string, credentialID string, username string) (string, time.Time, string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.ttl)
	sessionID := ids.New("SES")

	claims := Claims{
		PersonID:     personID,
		CredentialID: credentialID,
		Username:     username,
		SessionID:    sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionID,
			Subject:   personID,
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("sign jwt: %w", err)
	}

	return signed, expiresAt, sessionID, nil
}

func (m *JWTManager) Verify(rawToken string) (Claims, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Claims{}, ErrInvalidToken
	}

	claims := Claims{}

	parsed, err := jwt.ParseWithClaims(
		rawToken,
		&claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrUnexpectedSigningMethod
			}

			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithAudience(m.audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !parsed.Valid {
		return Claims{}, ErrInvalidToken
	}

	if claims.PersonID == "" || claims.CredentialID == "" || claims.Username == "" || claims.SessionID == "" {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}
