# TaifaID API

Base URL:

```text
http://localhost:8080
```

## Health

### `GET /healthz`

Returns service liveness.

Example response:

```json
{
  "correlation_id": "corr-...",
  "environment": "local",
  "service": "taifa-id",
  "status": "ok"
}
```

### `GET /readyz`

Returns dependency readiness.

Example response with database configured:

```json
{
  "correlation_id": "corr-...",
  "dependencies": {
    "database": "ok"
  },
  "service": "taifa-id",
  "status": "ok"
}
```

## Persons

### `POST /api/v1/persons`

Creates a synthetic person.

Request:

```json
{
  "synthetic_nin": "NIN-CIT-001",
  "display_name": "Amina Citizen"
}
```

Response:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "id": "PER-...",
    "synthetic_nin": "NIN-CIT-001",
    "display_name": "Amina Citizen",
    "status": "ACTIVE",
    "created_at": "2026-05-14T15:00:00Z",
    "updated_at": "2026-05-14T15:00:00Z"
  }
}
```

Errors:

```text
400 VALIDATION_ERROR
400 INVALID_JSON
409 CONFLICT
500 INTERNAL_ERROR
```

### `GET /api/v1/persons/{person_id}`

Returns a person by ID.

Example:

```powershell
Invoke-RestMethod http://localhost:8080/api/v1/persons/PER-...
```

Errors:

```text
400 VALIDATION_ERROR
404 NOT_FOUND
500 INTERNAL_ERROR
```

### `GET /api/v1/persons?synthetic_nin=...`

Returns a person by synthetic NIN.

Example:

```powershell
Invoke-RestMethod "http://localhost:8080/api/v1/persons?synthetic_nin=NIN-CIT-001"
```

Errors:

```text
400 VALIDATION_ERROR
404 NOT_FOUND
500 INTERNAL_ERROR
```

### `PATCH /api/v1/persons/{person_id}/status`

Updates a person status.

Request:

```json
{
  "status": "SUSPENDED"
}
```

Allowed statuses:

```text
ACTIVE
SUSPENDED
DECEASED
DUPLICATE_REVIEW
DISABLED
```

Errors:

```text
400 VALIDATION_ERROR
400 INVALID_JSON
404 NOT_FOUND
500 INTERNAL_ERROR
```

## Error Shape

All API errors use this shape:

```json
{
  "correlation_id": "corr-...",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Person request is invalid.",
    "correlation_id": "corr-..."
  }
}
```

## Audit Events

Person mutations write to `audit_outbox`.

Created person:

```text
identity.person.created
```

Status change:

```text
identity.person.status_changed
```

## Organizations

### `POST /api/v1/organizations`

Creates an organization.

Request:

```json
{
  "name": "Taifa Hospital",
  "primary_type": "HEALTH_PROVIDER"
}
```

Allowed primary types:

```text
GOVERNMENT_AGENCY
HEALTH_PROVIDER
EMPLOYER
FINANCIAL_INSTITUTION_SIM
SUPPLIER
```

Response:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "id": "ORG-...",
    "name": "Taifa Hospital",
    "primary_type": "HEALTH_PROVIDER",
    "status": "ACTIVE",
    "created_at": "2026-05-14T15:00:00Z",
    "updated_at": "2026-05-14T15:00:00Z"
  }
}
```

### `GET /api/v1/organizations`

Lists organizations.

### `GET /api/v1/organizations/{organization_id}`

Returns an organization by ID.

### `PATCH /api/v1/organizations/{organization_id}/status`

Updates an organization status.

Request:

```json
{
  "status": "SUSPENDED"
}
```

Allowed statuses:

```text
ACTIVE
SUSPENDED
DISABLED
```

### `POST /api/v1/organizations/{organization_id}/capabilities`

Adds an organization capability.

Request:

```json
{
  "capability": "CAN_PROVIDE_HEALTH_SERVICES"
}
```

Allowed capabilities:

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

### `GET /api/v1/organizations/{organization_id}/capabilities`

Lists capabilities for an organization.

### `DELETE /api/v1/organizations/{organization_id}/capabilities/{capability}`

Removes a capability from an organization.

Example:

```powershell
Invoke-RestMethod `
  -Method Delete `
  -Uri "http://localhost:8080/api/v1/organizations/ORG-.../capabilities/CAN_PROVIDE_HEALTH_SERVICES"
```

## Organization Audit Events

Organization mutations write to `audit_outbox`.

Created organization:

```text
identity.organization.created
```

Status changed:

```text
identity.organization.status_changed
```

Capability added:

```text
identity.organization_capability.added
```

Capability removed:

```text
identity.organization_capability.removed
```



## Memberships

### `POST /api/v1/memberships`

Creates an active organization membership.

The referenced person and organization must both exist and be `ACTIVE`.

Request:

```json
{
  "person_id": "PER-...",
  "organization_id": "ORG-...",
  "membership_type": "PROVIDER_STAFF"
}
```

Allowed membership types:

```text
EMPLOYEE
PROVIDER_STAFF
AGENCY_STAFF
FINANCIAL_OPERATOR
AUDITOR
SYSTEM_ADMIN
```

Response:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "id": "MEM-...",
    "person_id": "PER-...",
    "organization_id": "ORG-...",
    "membership_type": "PROVIDER_STAFF",
    "status": "ACTIVE",
    "starts_at": "2026-05-14T15:00:00Z",
    "ends_at": null,
    "created_at": "2026-05-14T15:00:00Z",
    "updated_at": "2026-05-14T15:00:00Z"
  }
}
```

### `GET /api/v1/memberships/{membership_id}`

Returns a membership by ID.

### `GET /api/v1/persons/{person_id}/memberships`

Lists memberships for a person.

### `GET /api/v1/organizations/{organization_id}/memberships`

Lists memberships for an organization.

### `PATCH /api/v1/memberships/{membership_id}/status`

Updates membership status.

Request:

```json
{
  "status": "SUSPENDED"
}
```

Allowed statuses:

```text
ACTIVE
SUSPENDED
ENDED
PENDING
```

When status is set to `ENDED`, `ends_at` is set if it was previously null.

### `POST /api/v1/memberships/{membership_id}/roles`

Adds a role to an active membership.

Request:

```json
{
  "role": "PROVIDER_CLINICIAN"
}
```

Allowed roles:

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

### `GET /api/v1/memberships/{membership_id}/roles`

Lists roles for a membership.

### `DELETE /api/v1/memberships/{membership_id}/roles/{role}`

Removes a role from a membership.

Example:

```powershell
Invoke-RestMethod `
  -Method Delete `
  -Uri "http://localhost:8080/api/v1/memberships/MEM-.../roles/PROVIDER_CLINICIAN"
```

## Membership Audit Events

Membership mutations write to `audit_outbox`.

Created membership:

```text
identity.membership.created
```

Status changed:

```text
identity.membership.status_changed
```

Role added:

```text
identity.membership_role.added
```

Role removed:

```text
identity.membership_role.removed
```


## Credentials

### `POST /api/v1/credentials`

Creates a credential for an existing active person.

The plaintext password is accepted only at creation time. It is hashed with bcrypt before storage and is never returned by the API.

Request:

```json
{
  "person_id": "PER-...",
  "username": "amina.provider",
  "password": "ExampleDevPass123!"
}
```

Rules:

```text
person_id must reference an existing ACTIVE person
username is normalized to lowercase
username must be unique
password must be 8 to 256 characters
```

Response:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "id": "CRD-...",
    "person_id": "PER-...",
    "username": "amina.provider",
    "status": "ACTIVE",
    "created_at": "2026-05-14T15:00:00Z",
    "updated_at": "2026-05-14T15:00:00Z"
  }
}
```

Errors:

```text
400 INVALID_JSON
400 VALIDATION_ERROR
404 NOT_FOUND
409 CONFLICT
500 INTERNAL_ERROR
```

### `GET /api/v1/credentials/{credential_id}`

Returns a credential without the password hash.

### `GET /api/v1/persons/{person_id}/credentials`

Lists credentials for a person without password hashes.

## Credential Audit Events

Credential creation writes to `audit_outbox`.

Created credential:

```text
identity.credential.created
```

## Authentication

### `POST /api/v1/auth/login`

Authenticates a credential and returns a JWT access token.

Request:

```json
{
  "username": "amina.credential",
  "password": "ExampleDevPass123!"
}
```

Response:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "access_token": "eyJ...",
    "token_type": "Bearer",
    "expires_at": "2026-05-14T18:00:00Z",
    "session_id": "SES-...",
    "person_id": "PER-...",
    "credential_id": "CRD-...",
    "username": "amina.credential"
  }
}
```

The JWT proves authentication only. It does not carry complete authorization truth.

Authorization remains split:

```text
TaifaID:
  authentication and actor identity

TaifaExchange:
  request-boundary authorization

Domain system:
  domain-specific authorization
```

### `POST /api/v1/auth/introspect`

Validates a JWT and checks that the referenced person and credential are still active.

Request:

```json
{
  "token": "eyJ..."
}
```

Response for an active token:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "active": true,
    "person_id": "PER-...",
    "credential_id": "CRD-...",
    "username": "amina.credential",
    "session_id": "SES-...",
    "issued_at": "2026-05-14T17:00:00Z",
    "expires_at": "2026-05-14T18:00:00Z"
  }
}
```

Response for an invalid, expired, disabled, or revoked-by-status token:

```json
{
  "correlation_id": "corr-...",
  "data": {
    "active": false
  }
}
```

## Authentication Audit Events

Successful login:

```text
identity.auth.succeeded
```

Failed login:

```text
identity.auth.failed
```

Token introspection:

```text
identity.auth.token_introspected
```
