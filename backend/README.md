# Backend (service-virtualization)

This folder contains the backend service and a self-contained Docker Compose for local development.

Quick start (from inside `backend/`):

```bash
# build and start postgres, localstack, and backend
docker compose up --build

# or run in background
docker compose up --build -d

# view logs
docker compose logs -f backend
```

Notes:
- Compose defines `localstack` (S3/SQS) and `postgres` and the `backend` service.
- The compose file is `backend/docker-compose.yml` and expects the Dockerfile in this folder.
- Environment variables are set in the compose file; override them with `-e` or an `.env` file if needed.
