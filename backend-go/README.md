# Service Virtualization - Go backend

This folder contains the Go backend split into three AWS Lambda functions. The implementation follows the existing Spring Boot behavior while separating control-plane management, capture, and runtime serving.

Lambda entry points:

- `cmd/management`: API registry, manual stubs, request history, versions, and analytics.
- `cmd/capture`: receives a target request, calls the target, writes bodies to S3, and writes metadata to Aurora/Postgres.
- `cmd/runtime`: receives application traffic such as `POST /claim`, checks virtualization, matches an enabled stub, reads an S3 response when configured, and returns it.
- `cmd/auth`: existing authentication endpoints.

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
- `S3_BUCKET` - S3 bucket for captured request and response bodies. If absent, local memory storage is used.
- `AWS_REGION` - AWS region used by the S3 client.
- `AWS_ENDPOINT` - optional LocalStack or compatible S3 endpoint.

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

The Serverless config is in `serverless.yml` and maps routes to separate functions:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/verify-token`
- `GET /api/admin-only` (RBAC-protected example)
- `/api/registry` and `/api/stubs` (management)
- `/api/capture` (capture)
- `/{proxy+}` (runtime virtualization)

## RBAC

This scaffold includes role-based access control (RBAC). When a user registers with `adGroup` set to `QED_DEV_OPS`, they receive the `Dev Ops` role; otherwise they get `Default User`.

The JWT contains `role` and `adGroup` claims and the code provides middleware to protect routes requiring specific roles. Example protected route: `GET /api/admin-only` requires the `Dev Ops` role.
