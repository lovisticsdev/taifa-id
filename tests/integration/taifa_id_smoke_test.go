package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type apiEnvelope struct {
	CorrelationID string          `json:"correlation_id"`
	Data          json.RawMessage `json:"data"`
	Error         *apiError       `json:"error"`
}

type apiError struct {
	Code          string `json:"code"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}

type loginResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
	SessionID    string    `json:"session_id"`
	PersonID     string    `json:"person_id"`
	CredentialID string    `json:"credential_id"`
	Username     string    `json:"username"`
}

type actorContextResponse struct {
	ActorContextID string `json:"actor_context_id"`
	PersonID       string `json:"person_id"`
	CredentialID   string `json:"credential_id"`
	Username       string `json:"username"`
	OrganizationID string `json:"organization_id"`
	Memberships    []struct {
		ID             string `json:"id"`
		MembershipType string `json:"membership_type"`
	} `json:"memberships"`
	Roles      []string   `json:"roles"`
	SessionID  string     `json:"session_id"`
	IssuedAt   *time.Time `json:"issued_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	ResolvedAt time.Time  `json:"resolved_at"`
}

func TestTaifaIDSmoke(t *testing.T) {
	baseURL := strings.TrimRight(os.Getenv("TAIFA_ID_TEST_BASE_URL"), "/")
	if baseURL == "" {
		t.Skip("TAIFA_ID_TEST_BASE_URL is not set")
	}

	username := envOrDefault("TAIFA_ID_TEST_USERNAME", "clinician.seed")
	password := envOrDefault("TAIFA_ID_TEST_PASSWORD", "ExampleDevPass123!")
	organizationID := envOrDefault("TAIFA_ID_TEST_ORGANIZATION_ID", "ORG-HP-CLINIC")
	expectedRole := envOrDefault("TAIFA_ID_TEST_EXPECTED_ROLE", "PROVIDER_CLINICIAN")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	health := getObject(t, client, baseURL+"/healthz")
	assertStringField(t, health, "status", "ok")

	ready := getObject(t, client, baseURL+"/readyz")
	assertStringField(t, ready, "status", "ok")
	assertNestedStringField(t, ready, "dependencies", "database", "ok")

	login := postEnvelope[loginResponse](
		t,
		client,
		baseURL+"/api/v1/auth/login",
		map[string]string{
			"username": username,
			"password": password,
		},
		http.StatusOK,
	)

	if login.AccessToken == "" {
		t.Fatal("login response access_token is empty")
	}

	if login.TokenType != "Bearer" {
		t.Fatalf("expected token_type Bearer, got %q", login.TokenType)
	}

	actorContext := postEnvelope[actorContextResponse](
		t,
		client,
		baseURL+"/api/v1/actor-context/resolve",
		map[string]string{
			"token":           login.AccessToken,
			"organization_id": organizationID,
		},
		http.StatusOK,
	)

	if actorContext.ActorContextID == "" {
		t.Fatal("actor context id is empty")
	}

	if actorContext.OrganizationID != organizationID {
		t.Fatalf("expected organization_id %q, got %q", organizationID, actorContext.OrganizationID)
	}

	if !contains(actorContext.Roles, expectedRole) {
		t.Fatalf("expected role %q in %v", expectedRole, actorContext.Roles)
	}
}

func getObject(t *testing.T, client *http.Client, url string) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create GET request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read GET response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned status %d: %s", url, resp.StatusCode, string(body))
	}

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode GET response JSON: %v\nbody: %s", err, string(body))
	}

	return decoded
}

func postEnvelope[T any](t *testing.T, client *http.Client, url string, request any, expectedStatus int) T {
	t.Helper()

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read POST response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("POST %s returned status %d, expected %d: %s", url, resp.StatusCode, expectedStatus, string(responseBody))
	}

	var envelope apiEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		t.Fatalf("decode response envelope: %v\nbody: %s", err, string(responseBody))
	}

	if envelope.Error != nil {
		t.Fatalf("unexpected API error: code=%s message=%s correlation_id=%s", envelope.Error.Code, envelope.Error.Message, envelope.Error.CorrelationID)
	}

	if len(envelope.Data) == 0 {
		t.Fatalf("response data is empty: %s", string(responseBody))
	}

	var data T
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode response data: %v\nbody: %s", err, string(responseBody))
	}

	return data
}

func assertStringField(t *testing.T, obj map[string]any, key string, expected string) {
	t.Helper()

	value, ok := obj[key].(string)
	if !ok {
		t.Fatalf("expected field %q to be string in %v", key, obj)
	}

	if value != expected {
		t.Fatalf("expected %s=%q, got %q", key, expected, value)
	}
}

func assertNestedStringField(t *testing.T, obj map[string]any, parent string, key string, expected string) {
	t.Helper()

	parentValue, ok := obj[parent].(map[string]any)
	if !ok {
		t.Fatalf("expected field %q to be object in %v", parent, obj)
	}

	value, ok := parentValue[key].(string)
	if !ok {
		t.Fatalf("expected field %q.%q to be string in %v", parent, key, obj)
	}

	if value != expected {
		t.Fatalf("expected %s.%s=%q, got %q", parent, key, expected, value)
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}

func Example_requiredEnvironment() {
	fmt.Println("TAIFA_ID_TEST_BASE_URL=http://localhost:8080")
	// Output:
	// TAIFA_ID_TEST_BASE_URL=http://localhost:8080
}
