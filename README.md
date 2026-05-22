# Service Virtualization Dashboard

A sample service virtualization dashboard with:
- React frontend for dashboard, API registration, stub management, and traffic records
- Java Spring Boot backend for API registration, proxying, and stub handling

Features:
- Register APIs to virtualize
- Toggle mock mode on/off per API
- Record proxied traffic and mock responses
- Manage request stubs for virtualized APIs
- Sidebar with Dashboard, APIs, Stubs, Records

## Run locally

### Backend
```bash
cd backend
./mvnw spring-boot:run
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

The frontend will proxy API requests to the backend for local development.
