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
