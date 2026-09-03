# Service Virtualization - Go backend (AWS Lambda auth scaffold)

This folder contains the Go backend for service virtualization, implemented as a single AWS Lambda function with an HTTP router.

The Lambda entrypoint is `cmd/auth/main.go`, and the API Gateway routes are defined in `serverless.yml`.

## Setup

```bash
cd backend-go
go mod tidy
```

## Local validation

To confirm the Go project still builds:

```bash
go build ./...
```

## Environment variables

- `JWT_SECRET` - secret used to sign tokens (default: `change-this-secret`)
- `DATABASE_URL` - optional Postgres connection string. Example: `postgres://user:pass@host:5432/dbname`

## Lambda / API Gateway deployment

Requires: `serverless` CLI and `serverless-plugin-go-build`.

```bash
# install Serverless Framework (npm)
npm install -g serverless

# deploy (ensure AWS credentials are configured)
cd backend-go
export JWT_SECRET="your-production-secret"
serverless deploy
```

The Serverless config is in `serverless.yml` and maps the following routes to the same Lambda:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/verify-token`
- `GET /api/admin-only` (RBAC-protected example)

## RBAC

This scaffold includes role-based access control (RBAC). When a user registers with `adGroup` set to `QED_DEV_OPS`, they receive the `Dev Ops` role; otherwise they get `Default User`.

The JWT contains `role` and `adGroup` claims and the code provides middleware to protect routes requiring specific roles. Example protected route: `GET /api/admin-only` requires the `Dev Ops` role.
