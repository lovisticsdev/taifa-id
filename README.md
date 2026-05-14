# TaifaID

TaifaID is the identity and registry service for the Taifa Republic Digital Public Infrastructure Range.

It is intentionally synthetic. It is designed for secure software engineering, red-team validation, telemetry, auditability, forensics, remediation, and reporting without using real citizen data or real production secrets.

## Doctrine

```text
Build defensively first.
Validate offensively second.
Remediate and retest third.
Report professionally throughout.
```

## Ownership boundary

TaifaID owns identity and registry truth.

It owns:

```text
person registry truth
organization registry truth
organization capability truth
organization membership truth
membership role truth
credential truth
authentication token issuance
token introspection
organization-scoped actor-context resolution
identity audit evidence
```

It does not own:

```text
route-boundary authorization
domain authorization
care claims
tax contributions
payment instructions
settlement
observability cases
citizen-facing service requests
```

Those responsibilities belong to other Taifa Republic systems.

```text
TaifaExchange owns exchange, routing, policy, and request-boundary truth.
TaifaObserve owns audit, evidence, detection, case, and timeline truth.
TaifaCare owns health/UHC care-financing truth.
TaifaTax owns contribution truth.
TaifaPay owns payment-instruction, synthetic switch, reserve-position, settlement, payout-status, and reconciliation truth.
TaifaCitizen owns citizen-facing view and service-request truth.
```

Core invariant:

```text
Every meaningful state transition must be owned, authorized, correlated, observable, investigable, retestable, and reportable.
```

## Current implementation status

Implemented:

```text
HTTP service scaffold
health and readiness probes
PostgreSQL connection pool
schema migrations
audit outbox
person registry
organization registry
organization capabilities
organization memberships
membership roles
credentials with bcrypt password hashing
JWT authentication
token introspection
actor-context resolution
canonical seed-data command
```

Not yet implemented:

```text
Docker runtime hardening
Docker Compose local full-stack workflow
black-box integration smoke tests
OpenAPI generation
authorization middleware
rate limiting
MFA
refresh tokens
credential lockout policy
outbox dispatcher
OpenTelemetry export
admin authorization surface
```

## Tech stack

```text
Go
PostgreSQL
pgx
chi
bcrypt
JWT HS256
Docker, optional
AWS RDS PostgreSQL, optional
```

The module is:

```text
taifa-id
```

Primary commands:

```text
cmd/taifa-id
cmd/taifa-id-seed
```

## Repository layout

```text
taifa-id/
  cmd/
    taifa-id/
    taifa-id-seed/

  internal/
    actorcontext/
    app/
    audit/
    auth/
    config/
    credential/
    membership/
    organization/
    person/
    platform/
    seed/

  migrations/
    000001_init.up.sql
    000001_init.down.sql
    000002_add_observe_security_events_capability.up.sql
    000002_add_observe_security_events_capability.down.sql

  tests/
    integration/

  .env.example
  .gitignore
  .gitattributes
  .dockerignore
  Dockerfile
  docker-compose.yml
  go.mod
  go.sum
  README.md
```

This repository should use one comprehensive README for now. Avoid a separate `docs/` folder until the documentation becomes large enough to justify splitting.

## Environment variables

Core service configuration:

```text
TAIFA_ID_SERVICE_NAME
TAIFA_ID_ENV

TAIFA_ID_HTTP_ADDR
TAIFA_ID_HTTP_READ_TIMEOUT
TAIFA_ID_HTTP_WRITE_TIMEOUT
TAIFA_ID_HTTP_IDLE_TIMEOUT
TAIFA_ID_HTTP_SHUTDOWN_TIMEOUT

TAIFA_ID_DATABASE_DSN
TAIFA_ID_DATABASE_MIN_CONNS
TAIFA_ID_DATABASE_MAX_CONNS
TAIFA_ID_DATABASE_CONNECT_TIMEOUT

TAIFA_ID_JWT_SECRET
TAIFA_ID_JWT_ISSUER
TAIFA_ID_JWT_AUDIENCE
TAIFA_ID_JWT_TTL

TAIFA_ID_SEED_PASSWORD
TAIFA_ID_SEED_TIMEOUT

TAIFA_ID_TEST_BASE_URL
TAIFA_ID_TEST_USERNAME
TAIFA_ID_TEST_PASSWORD
TAIFA_ID_TEST_ORGANIZATION_ID
TAIFA_ID_TEST_EXPECTED_ROLE
```

Recommended local PowerShell setup:

```powershell
$env:TAIFA_ID_DATABASE_CONNECT_TIMEOUT="15s"
$env:TAIFA_ID_DATABASE_DSN="host=localhost port=5432 user=taifa password=taifa_dev_password dbname=taifa_id sslmode=disable"

$env:TAIFA_ID_JWT_SECRET="replace-this-with-a-long-local-dev-secret-at-least-32-chars"
$env:TAIFA_ID_JWT_ISSUER="taifa-id"
$env:TAIFA_ID_JWT_AUDIENCE="taifa-republic"
$env:TAIFA_ID_JWT_TTL="1h"

$env:TAIFA_ID_SEED_TIMEOUT="5m"
```

Recommended AWS RDS PowerShell setup:

```powershell
$env:TAIFA_ID_DATABASE_CONNECT_TIMEOUT="15s"
$env:TAIFA_ID_DATABASE_DSN="host=$env:TAIFA_ID_DB_HOST port=5432 user=taifa password=$env:TAIFA_ID_DB_PASSWORD dbname=taifa_id sslmode=require"

$env:TAIFA_ID_JWT_SECRET="replace-this-with-a-long-local-dev-secret-at-least-32-chars"
$env:TAIFA_ID_JWT_ISSUER="taifa-id"
$env:TAIFA_ID_JWT_AUDIENCE="taifa-republic"
$env:TAIFA_ID_JWT_TTL="1h"
```

Do not commit real RDS credentials, JWT secrets, access tokens, or `.env` files.

## Database

The service expects PostgreSQL.

The initial schema creates:

```text
persons
organizations
organization_capabilities
organization_memberships
membership_roles
credentials
audit_outbox
```

Apply migrations to local Docker PostgreSQL:

```powershell
Get-Content migrations\000001_init.up.sql | docker compose exec -T postgres psql -U taifa -d taifa_id
Get-Content migrations\000002_add_observe_security_events_capability.up.sql | docker compose exec -T postgres psql -U taifa -d taifa_id
```

Apply migrations to AWS RDS:

```powershell
Get-Content migrations\000001_init.up.sql | docker run -i --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require"

Get-Content migrations\000002_add_observe_security_events_capability.up.sql | docker run -i --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require"
```

Verify tables:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require" `
  -c "\dt"
```

Expected tables:

```text
audit_outbox
credentials
membership_roles
organization_capabilities
organization_memberships
organizations
persons
```

## Run locally

Run compile checks:

```powershell
go fmt ./...
go mod tidy
go test ./...
```

Run the service:

```powershell
go run ./cmd/taifa-id
```

Expected log:

```text
starting HTTP server
```

Health check:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
```

Readiness check:

```powershell
Invoke-RestMethod http://localhost:8080/readyz
```

Expected readiness:

```text
database=ok
```

## Seed data

Run the seed command after migrations have been applied:

```powershell
$env:TAIFA_ID_SEED_TIMEOUT="5m"
go run ./cmd/taifa-id-seed
```

The seed command is idempotent.

Expected first complete run on an empty database:

```text
persons=8
organizations=9
capabilities=16
memberships=8
roles=9
credentials=8
audit_events=58
```

Expected repeated run:

```text
persons=0
organizations=0
capabilities=0
memberships=0
roles=0
credentials=0
audit_events=0
```

Default local seed password:

```text
ExampleDevPass123!
```

Override it with:

```powershell
$env:TAIFA_ID_SEED_PASSWORD="your-local-dev-password"
```

The seed command should not print the password value.

### Seed persons and credentials

| Person ID                  | Username           | Display name         |
| -------------------------- | ------------------ | -------------------- |
| `PER-SEED-CITIZEN-001`   | `citizen.seed`   | Amina Citizen Seed   |
| `PER-SEED-CLINICIAN-001` | `clinician.seed` | Nia Clinician Seed   |
| `PER-SEED-CLAIMS-001`    | `claims.seed`    | Peter Claims Seed    |
| `PER-SEED-EMPLOYER-001`  | `employer.seed`  | Esther Employer Seed |
| `PER-SEED-TAX-001`       | `tax.seed`       | Omar Tax Seed        |
| `PER-SEED-PAY-001`       | `pay.seed`       | Grace Pay Seed       |
| `PER-SEED-OBSERVE-001`   | `observe.seed`   | David Observe Seed   |
| `PER-SEED-ADMIN-001`     | `admin.seed`     | Sana Admin Seed      |

### Seed organizations

| Organization ID     | Primary type                  | Name                             |
| ------------------- | ----------------------------- | -------------------------------- |
| `ORG-GOV-CARE`    | `GOVERNMENT_AGENCY`         | Taifa Care Authority             |
| `ORG-GOV-TAX`     | `GOVERNMENT_AGENCY`         | Taifa Tax Authority              |
| `ORG-GOV-PAY`     | `GOVERNMENT_AGENCY`         | Taifa Pay Authority              |
| `ORG-GOV-OBSERVE` | `GOVERNMENT_AGENCY`         | Taifa Observe Authority          |
| `ORG-HP-HOSP`     | `HEALTH_PROVIDER`           | Taifa National Hospital          |
| `ORG-HP-CLINIC`   | `HEALTH_PROVIDER`           | Taifa Community Clinic           |
| `ORG-EMP-MFG`     | `EMPLOYER`                  | Taifa Manufacturing Employer     |
| `ORG-FIN-CB`      | `FINANCIAL_INSTITUTION_SIM` | Taifa Central Bank Simulation    |
| `ORG-FIN-COMM`    | `FINANCIAL_INSTITUTION_SIM` | Taifa Commercial Bank Simulation |

### Useful seed actor contexts

| Username           | Organization ID     | Membership type        | Role                                     |
| ------------------ | ------------------- | ---------------------- | ---------------------------------------- |
| `clinician.seed` | `ORG-HP-CLINIC`   | `PROVIDER_STAFF`     | `PROVIDER_CLINICIAN`                   |
| `claims.seed`    | `ORG-HP-HOSP`     | `PROVIDER_STAFF`     | `PROVIDER_CLAIMS_OFFICER`              |
| `employer.seed`  | `ORG-EMP-MFG`     | `EMPLOYEE`           | `EMPLOYER_SUBMITTER`                   |
| `tax.seed`       | `ORG-GOV-TAX`     | `AGENCY_STAFF`       | `TAX_OFFICER`                          |
| `pay.seed`       | `ORG-GOV-PAY`     | `FINANCIAL_OPERATOR` | `PAY_OPERATOR`                         |
| `observe.seed`   | `ORG-GOV-OBSERVE` | `AUDITOR`            | `OBSERVE_ANALYST`, `OBSERVE_AUDITOR` |
| `admin.seed`     | `ORG-GOV-OBSERVE` | `SYSTEM_ADMIN`       | `SYSTEM_ADMIN`                         |

Verify seed credential count:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require" `
  -c "select count(*) as seed_credentials from credentials where id like 'CRD-SEED-%';"
```

Expected:

```text
8
```

Verify seed credential audit coverage:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require" `
  -c "select count(*) as seed_credential_audits from audit_outbox where correlation_id = 'seed' and event_type = 'identity.credential.created' and resource_id like 'CRD-SEED-%';"
```

Expected:

```text
8
```

## API surface

All successful responses include a top-level `correlation_id` and `data`.

All error responses include a top-level `correlation_id` and `error`.

Common error codes:

```text
INVALID_JSON
VALIDATION_ERROR
NOT_FOUND
CONFLICT
UNAUTHORIZED
FORBIDDEN
INTERNAL_ERROR
```

### Health

```text
GET /healthz
GET /readyz
```

### Persons

```text
POST  /api/v1/persons
GET   /api/v1/persons/{person_id}
GET   /api/v1/persons?synthetic_nin=...
PATCH /api/v1/persons/{person_id}/status
```

Create person:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/persons `
  -ContentType "application/json" `
  -Body '{"synthetic_nin":"NIN-CIT-001","display_name":"Amina Citizen"}'
```

Valid person statuses:

```text
ACTIVE
SUSPENDED
DECEASED
MERGED
```

Audit events:

```text
identity.person.created
identity.person.status_changed
```

### Organizations

```text
POST  /api/v1/organizations
GET   /api/v1/organizations
GET   /api/v1/organizations/{organization_id}
PATCH /api/v1/organizations/{organization_id}/status
```

Create organization:

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/organizations `
  -ContentType "application/json" `
  -Body '{"name":"Taifa Community Clinic","primary_type":"HEALTH_PROVIDER"}'
```

Valid primary types:

```text
GOVERNMENT_AGENCY
HEALTH_PROVIDER
EMPLOYER
FINANCIAL_INSTITUTION_SIM
SUPPLIER
```

Valid organization statuses:

```text
ACTIVE
SUSPENDED
CLOSED
```

Audit events:

```text
identity.organization.created
identity.organization.status_changed
```

### Organization capabilities

```text
POST   /api/v1/organizations/{organization_id}/capabilities
GET    /api/v1/organizations/{organization_id}/capabilities
DELETE /api/v1/organizations/{organization_id}/capabilities/{capability}
```

Valid capabilities:

```text
CAN_EMPLOY_PERSONS
CAN_SUBMIT_TAX_CONTRIBUTIONS
CAN_PROVIDE_HEALTH_SERVICES
CAN_RECEIVE_HEALTH_PAYOUTS
CAN_ROUTE_PAYMENTS
CAN_HOLD_RESERVE_ACCOUNT
CAN_OPERATE_GOVERNMENT_SERVICE
CAN_OBSERVE_SECURITY_EVENTS
```

Audit events:

```text
identity.organization_capability.added
identity.organization_capability.removed
```

### Memberships

```text
POST  /api/v1/memberships
GET   /api/v1/memberships/{membership_id}
GET   /api/v1/persons/{person_id}/memberships
GET   /api/v1/organizations/{organization_id}/memberships
PATCH /api/v1/memberships/{membership_id}/status
```

Create membership:

```powershell
$membershipBody = @{
  person_id = "PER-..."
  organization_id = "ORG-..."
  membership_type = "PROVIDER_STAFF"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/memberships `
  -ContentType "application/json" `
  -Body $membershipBody
```

Valid membership types:

```text
EMPLOYEE
PROVIDER_STAFF
AGENCY_STAFF
FINANCIAL_OPERATOR
AUDITOR
SYSTEM_ADMIN
```

Valid membership statuses:

```text
ACTIVE
SUSPENDED
ENDED
```

Rules:

```text
person must exist and be ACTIVE
organization must exist and be ACTIVE
setting status to ENDED records ends_at
```

Audit events:

```text
identity.membership.created
identity.membership.status_changed
```

### Membership roles

```text
POST   /api/v1/memberships/{membership_id}/roles
GET    /api/v1/memberships/{membership_id}/roles
DELETE /api/v1/memberships/{membership_id}/roles/{role}
```

Valid roles:

```text
CITIZEN
PROVIDER_CLINICIAN
PROVIDER_CLAIMS_OFFICER
CARE_ADJUDICATOR
TAX_OFFICER
EMPLOYER_SUBMITTER
PAY_OPERATOR
OBSERVE_ANALYST
OBSERVE_AUDITOR
SYSTEM_ADMIN
```

Audit events:

```text
identity.membership_role.added
identity.membership_role.removed
```

### Credentials

```text
POST /api/v1/credentials
GET  /api/v1/credentials/{credential_id}
GET  /api/v1/persons/{person_id}/credentials
```

Create credential:

```powershell
$credentialBody = @{
  person_id = "PER-..."
  username = "amina.seed"
  password = "ExampleDevPass123!"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/credentials `
  -ContentType "application/json" `
  -Body $credentialBody
```

Credential rules:

```text
person must exist and be ACTIVE
username is normalized to lowercase
username must be unique
password length must be 8 to 256 characters
password is stored as a bcrypt hash
password_hash is never returned by the API
```

Credential statuses:

```text
ACTIVE
DISABLED
LOCKED
```

Audit event:

```text
identity.credential.created
```

### Authentication

```text
POST /api/v1/auth/login
POST /api/v1/auth/introspect
```

Login:

```powershell
$login = Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"clinician.seed","password":"ExampleDevPass123!"}'

$token = $login.data.access_token
```

Login response data:

```json
{
  "access_token": "eyJ...",
  "token_type": "Bearer",
  "expires_at": "2026-05-14T18:00:00Z",
  "session_id": "SES-...",
  "person_id": "PER-...",
  "credential_id": "CRD-...",
  "username": "clinician.seed"
}
```

Token introspection:

```powershell
$introspectBody = @{
  token = $token
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/auth/introspect `
  -ContentType "application/json" `
  -Body $introspectBody
```

Audit events:

```text
identity.auth.succeeded
identity.auth.failed
identity.auth.token_introspected
```

### Actor context

POST /api/v1/actor-context/resolve

Actor-context resolution turns a valid authenticated person into an organization-scoped actor context.

Request:

```powershell
$actorContextBody = @{
  token = $token
  organization_id = "ORG-HP-CLINIC"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/actor-context/resolve `
  -ContentType "application/json" `
  -Body $actorContextBody
```

Resolution checks:

```text
token signature is valid
token issuer and audience are valid
token is not expired
credential exists and is ACTIVE
person exists and is ACTIVE
organization exists and is ACTIVE
person has at least one ACTIVE membership in the requested organization
roles are collected from active memberships in that organization
```

Response data:

```json
{
  "actor_context_id": "ACTX-...",
  "person_id": "PER-SEED-CLINICIAN-001",
  "credential_id": "CRD-SEED-CLINICIAN",
  "username": "clinician.seed",
  "organization_id": "ORG-HP-CLINIC",
  "memberships": [
    {
      "id": "MEM-SEED-CLINICIAN-CLINIC",
      "membership_type": "PROVIDER_STAFF"
    }
  ],
  "roles": [
    "PROVIDER_CLINICIAN"
  ],
  "session_id": "SES-...",
  "issued_at": "2026-05-14T17:00:00Z",
  "expires_at": "2026-05-14T18:00:00Z",
  "resolved_at": "2026-05-14T17:05:00Z"
}
```

Audit events:

```text
identity.actor_context.allowed
identity.actor_context.denied
```

Important boundary:

```text
A JWT proves authentication.
An actor context proves current organization-scoped identity context.
TaifaExchange will own route-boundary authorization.
Domain systems will own domain authorization.
```

## End-to-end seeded test

Start the service in one terminal:

```powershell
go run ./cmd/taifa-id
```

In another terminal:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/readyz

$login = Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"clinician.seed","password":"ExampleDevPass123!"}'

$token = $login.data.access_token

$actorContextBody = @{
  token = $token
  organization_id = "ORG-HP-CLINIC"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8080/api/v1/actor-context/resolve `
  -ContentType "application/json" `
  -Body $actorContextBody
```

Expected actor-context facts:

```text
organization_id = ORG-HP-CLINIC
membership_type = PROVIDER_STAFF
role = PROVIDER_CLINICIAN
```

## Integration smoke test

The black-box integration test should live under:

```text
tests/integration/taifa_id_smoke_test.go
```

It should skip automatically unless `TAIFA_ID_TEST_BASE_URL` is set.

Run it against a running, migrated, seeded service:

```powershell
$env:TAIFA_ID_TEST_BASE_URL="http://localhost:8080"
$env:TAIFA_ID_TEST_USERNAME="clinician.seed"
$env:TAIFA_ID_TEST_PASSWORD="ExampleDevPass123!"
$env:TAIFA_ID_TEST_ORGANIZATION_ID="ORG-HP-CLINIC"
$env:TAIFA_ID_TEST_EXPECTED_ROLE="PROVIDER_CLINICIAN"

go test ./tests/integration
```

Smoke-test scope:

```text
GET /healthz
GET /readyz
POST /api/v1/auth/login
POST /api/v1/actor-context/resolve
```

Normal package checks should still work without the service running:

```powershell
go test ./...
```

## Docker

Docker support should be kept minimal.

Expected files:

```text
Dockerfile
.dockerignore
docker-compose.yml
```

The Docker image should build the main service binary only:

```text
cmd/taifa-id
```

The seed command should remain an explicit operator action:

```powershell
go run ./cmd/taifa-id-seed
```

Build image:

```powershell
docker build -t taifa-id:local .
```

Run local PostgreSQL with Compose:

```powershell
docker compose up -d postgres
```

Run service container after migrations:

```powershell
docker compose --profile app up --build taifa-id
```

Do not place real RDS credentials in `docker-compose.yml`.

## Audit model

Audit events are written to `audit_outbox`.

The service writes audit evidence for:

```text
person create/status change
organization create/status change
organization capability add/remove
membership create/status change
membership role add/remove
credential create
auth success/failure
token introspection
actor-context allow/deny
```

Check latest audit events:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_ID_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_ID_DB_HOST port=5432 dbname=taifa_id user=taifa sslmode=require" `
  -c "select event_type, resource_type, resource_id, result, correlation_id from audit_outbox order by created_at desc limit 20;"
```

## Security notes

This repository is for synthetic data only.

Never commit:

```text
real citizen data
real health data
real tax data
real payment data
RDS passwords
JWT secrets
access tokens
.env files
terminal transcripts containing secrets
```

If a secret is exposed in chat, screenshots, logs, or commits, rotate it.

Password handling:

```text
passwords are hashed with bcrypt
password_hash is never returned by API responses
seed passwords are local development credentials only
```

JWT handling:

```text
JWTs use HS256
TAIFA_ID_JWT_SECRET must be at least 32 characters
tokens are bearer credentials
token possession is not authorization
actor-context resolution is required for organization-scoped operation
```

AWS RDS development posture:

```text
public access may be acceptable only for dev
PostgreSQL inbound access should be restricted to current public IP /32
sslmode=require should be used
RDS credentials should be rotated after any exposure
```

## Troubleshooting

### `No connection could be made because the target machine actively refused it`

The API server is not running on `localhost:8080`.

Start it:

```powershell
go run ./cmd/taifa-id
```

### `/readyz` shows `database=unavailable`

Check:

```text
TAIFA_ID_DATABASE_DSN
database host
database port
database credentials
RDS security group
current public IP allowlist
sslmode
```

### `open postgres: ping postgres: context deadline exceeded`

Usually database network reachability or slow RDS connection.

Increase timeout:

```powershell
$env:TAIFA_ID_DATABASE_CONNECT_TIMEOUT="15s"
```

Then retry.

### `initialize jwt manager: invalid jwt config`

`TAIFA_ID_JWT_SECRET` is missing or too short.

Set a local secret of at least 32 characters:

```powershell
$env:TAIFA_ID_JWT_SECRET="replace-this-with-a-long-local-dev-secret-at-least-32-chars"
```

### `Username already exists`

The test user or seed user already exists.

Use a different username for manual tests, or rely on the idempotent seed records.

### Seed command times out

Use a longer seed timeout:

```powershell
$env:TAIFA_ID_SEED_TIMEOUT="5m"
```

Then rerun:

```powershell
go run ./cmd/taifa-id-seed
```

### Integration test skips

Set:

```powershell
$env:TAIFA_ID_TEST_BASE_URL="http://localhost:8080"
```

Then rerun:

```powershell
go test ./tests/integration
```

## Recommended pre-commit validation

```powershell
go fmt ./...
go mod tidy
go test ./...

$env:TAIFA_ID_TEST_BASE_URL="http://localhost:8080"
go test ./tests/integration

git status
```

## Current checkpoint

```text
Batch 0: service skeleton                           done
Batch 1: bootable HTTP service scaffold             done
Batch 2: PostgreSQL schema and audit outbox          done
Batch 3: person registry API                         done
Batch 4: organization registry and capabilities      done
Batch 5: membership registry API                     done
Batch 6: credential registry API                     done
Batch 7: authentication and JWT introspection        done
Batch 8: actor context resolution                    done
Batch 9: canonical seed data command                 done
Batch 9.1: seed command hardening                    pending/optional
Batch 10: slim repo stabilization                    in progress
```
