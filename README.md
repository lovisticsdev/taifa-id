
# TaifaID

TaifaID is the identity and registry authority for the Taifa Republic digital public infrastructure range.

It owns:

- synthetic person registry
- organization registry
- organization capabilities
- organization memberships
- membership roles
- credentials
- actor-context resolution
- identity audit events

## Current Implementation State

Implemented:

- config loading from environment variables
- HTTP server bootstrap
- request correlation IDs
- JSON response helpers
- error response helpers
- request logging middleware
- panic recovery middleware
- graceful shutdown
- `/healthz`
- `/readyz`
- PostgreSQL connection support through `TAIFA_ID_DATABASE_DSN`
- initial PostgreSQL schema migration
- audit outbox event model
- audit outbox repository
- prefixed ID helpers

Not implemented yet:

- person registry API
- organization registry API
- membership API
- credential API
- authentication API
- actor-context resolution API
- seed command
- integration tests
- audit outbox publisher

## Design Rule

TaifaID resolves identity and actor context.

It does not own:

- health coverage
- claims
- tax compliance
- payment settlement
- SOC cases
- domain authorization decisions

Authorization is split as follows:

```text
TaifaID:
  Who is this person?
  Which organization are they attached to?
  Which roles do they have in that organization?
  Is the actor context valid?

TaifaExchange:
  May this actor context call this route?

Domain system:
  Is this specific domain action valid?
```

## Run Locally Without Database

Install dependencies:

```powershell
go mod tidy
```

Start the service:

```powershell
go run ./cmd/taifa-id
```

Health check:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
```

Readiness check:

```powershell
Invoke-RestMethod http://localhost:8080/readyz
```

Expected readiness response when no database DSN is configured:

```json
{
  "correlation_id": "corr-...",
  "dependencies": {
    "database": "not_configured"
  },
  "service": "taifa-id",
  "status": "ok"
}
```

## PostgreSQL Configuration

TaifaID reads the database connection string from:

```env
TAIFA_ID_DATABASE_DSN
```

The service starts without a database DSN, but `/readyz` reports:

```text
database = not_configured
```

When a valid database DSN is configured, `/readyz` should report:

```text
database = ok
```

## AWS RDS Dev Database

For the current dev environment, TaifaID can connect to AWS RDS PostgreSQL.

Recommended dev RDS posture:

```text
DB identifier: taifa-id-db
Database name: taifa_id
Master username: taifa
Public access: yes, for local laptop development only
Security group inbound: PostgreSQL TCP 5432 from your public IP /32 only
SSL mode: require
```

Do not allow inbound PostgreSQL from:

```text
0.0.0.0/0
```

Set the DSN in PowerShell:

```powershell
$env:TAIFA_ID_DATABASE_DSN="postgres://taifa:<PASSWORD>@<RDS_ENDPOINT>:5432/taifa_id?sslmode=require"
```

Do not commit the real password or full real DSN.

## Apply Migration to AWS RDS

Using Docker:

```powershell
Get-Content migrations\000001_init.up.sql | docker run -i --rm postgres:16 `
  psql "postgres://taifa:<PASSWORD>@<RDS_ENDPOINT>:5432/taifa_id?sslmode=require"
```

Verify tables:

```powershell
docker run -it --rm postgres:16 `
  psql "postgres://taifa:<PASSWORD>@<RDS_ENDPOINT>:5432/taifa_id?sslmode=require" `
  -c "\dt"
```

Expected tables:

```text
persons
organizations
organization_capabilities
organization_memberships
membership_roles
credentials
audit_outbox
```

## Local Docker PostgreSQL Alternative

A local PostgreSQL container can also be used:

```powershell
docker run --name taifa-id-postgres `
  -e POSTGRES_USER=taifa `
  -e POSTGRES_PASSWORD=taifa `
  -e POSTGRES_DB=taifa_id `
  -p 5432:5432 `
  -d postgres:16
```

Set the local DSN:

```powershell
$env:TAIFA_ID_DATABASE_DSN="postgres://taifa:taifa@localhost:5432/taifa_id?sslmode=disable"
```

Apply migration:

```powershell
Get-Content migrations\000001_init.up.sql | docker exec -i taifa-id-postgres psql -U taifa -d taifa_id
```

## Validate

Run:

```powershell
go fmt ./...
go mod tidy
go test ./...
```

Start the service:

```powershell
go run ./cmd/taifa-id
```

Check readiness:

```powershell
Invoke-RestMethod http://localhost:8080/readyz
```

Expected with AWS RDS or local PostgreSQL configured:

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

## Environment Variables

```env
TAIFA_ID_SERVICE_NAME=taifa-id
TAIFA_ID_ENV=local

TAIFA_ID_HTTP_ADDR=:8080
TAIFA_ID_HTTP_READ_TIMEOUT=5s
TAIFA_ID_HTTP_WRITE_TIMEOUT=10s
TAIFA_ID_HTTP_IDLE_TIMEOUT=60s
TAIFA_ID_HTTP_SHUTDOWN_TIMEOUT=10s

TAIFA_ID_DATABASE_DSN=
TAIFA_ID_DATABASE_MIN_CONNS=1
TAIFA_ID_DATABASE_MAX_CONNS=5
TAIFA_ID_DATABASE_CONNECT_TIMEOUT=5s
```

## Batch Status

```text
Batch 0: service skeleton                         done
Batch 1: bootable HTTP service scaffold           done
Batch 2: PostgreSQL schema and audit outbox        in progress / validation pending
```




| Setting                | Recommended value                                               |
| ---------------------- | --------------------------------------------------------------- |
| Creation method        | Full configuration                                              |
| Template               | Free tier                                                       |
| Deployment             | Single-AZ DB instance                                           |
| Engine                 | PostgreSQL                                                      |
| Engine version         | Default selected version is fine                                |
| DB instance identifier | `taifa-id-db`                                                 |
| Master username        | `taifa`                                                       |
| Credentials            | Self managed is fine for now                                    |
| Password               | Use a strong password; do not commit it                         |
| Instance type          | `db.t4g.micro`                                                |
| Storage                | General Purpose SSD,`20 GiB`                                  |
| Storage autoscaling    | **Disable**for now                                        |
| Public access          | **Yes** , because your Go app is running from your laptop |
| VPC security group     | Create new, not default                                         |
| Inbound rule           | PostgreSQL `5432`from**your current public IP only**    |
| Initial database name  | `taifa_id`                                                    |
| Backup retention       | `1 day`                                                       |
| Encryption             | Enabled                                                         |
| Deletion protection    | Off for dev                                                     |
