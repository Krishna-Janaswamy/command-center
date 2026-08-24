# Service Virtualization — How to run (Frontend & Backend)

This document describes how to run the backend and frontend locally for development.

## Prerequisites
- Java (JDK 17 or later)
- Maven
- Node.js (18+) and npm or yarn
- Optional: Docker / LocalStack when testing AWS integrations

---

## Backend (Spring Boot)

- Defaults:
  - Port: `3001` (configurable in `backend/src/main/resources/application.properties`)
  - Database: SQLite file at `backend/data/data.db` (configurable via `db.path`)

- Recommended local env to avoid calling real AWS services:
```bash
export AWS_SQS_ENABLED=false
export AWS_STORAGE_ENABLED=false
```

- Run in development (uses Maven):
```bash
cd backend
mvn -DskipTests spring-boot:run
```

- Build and run jar:
```bash
cd backend
mvn -DskipTests package
java -jar target/service-virtualization-0.0.1-SNAPSHOT.jar
```

- If you need AWS or LocalStack for testing, set credentials or endpoint before starting:
```bash
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_ENDPOINT=http://localhost:4566   # LocalStack
export AWS_SQS_ENABLED=true
export AWS_STORAGE_ENABLED=true
```

---

## Frontend (Vite / React)

- Run development server:
```bash
cd frontend
npm install
npm run dev
```

- Build production bundle:
```bash
cd frontend
npm run build
# optionally serve the `dist` folder for a quick check
npx serve dist
```

---

## Full local workflow
1. Start the backend first (see Backend section). Confirm it listens on port `3001`.
2. Start the frontend. It will connect to the backend for API calls.

If you need the frontend to proxy API calls to a custom backend port, update the frontend config or the request code in `frontend/src/services/registryApi.js`.

## Troubleshooting
- If Spring Boot fails with `Port 3001 already in use`, find and kill the process:
```bash
lsof -nP -iTCP:3001 -sTCP:LISTEN
# then
kill <PID>
```
- If the app tries to access AWS and you don't want that, ensure the `AWS_*` env vars above are set to `false`.
- If database inserts fail due to missing columns (SQLite), the app expects a `requests` table with columns including `bodyS3Key` and `responseS3Key`. The default DB file is `backend/data/data.db`.

---