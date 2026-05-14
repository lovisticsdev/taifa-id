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
