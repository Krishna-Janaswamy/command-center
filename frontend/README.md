# Frontend (service-virtualization-dashboard)

This folder contains the frontend app (Vite + React) and a minimal Docker Compose to run the dev server.

Quick start (from inside `frontend/`):

```bash
docker compose up --build

# open http://localhost:5173
```

Notes:
- The frontend compose only builds and runs the frontend dev server.
- To connect to a locally running backend, set `VITE_BACKEND_URL` when running compose, for example:

```bash
VITE_BACKEND_URL=http://host.docker.internal:3001 docker compose up --build
```
