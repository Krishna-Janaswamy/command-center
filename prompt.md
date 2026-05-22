Build a Service Virtualization Platform ("Command Center") with Spring Boot 3 (Java 21) backend and React 18 frontend. Two feature areas: Health Monitoring and Service Virtualization, with JWT auth and RBAC.
Tech Stack — Backend: Spring Boot 3.3.x, Java 21, SQLite (sqlite-jdbc + HikariCP, single-connection pool), JDBC Template (no JPA), Custom HMAC-SHA256 JWT (no Spring Security OAuth), Maven.
Frontend: React 18, React Router v6, Axios, Chart.js/react-chartjs-2, CRA with http-proxy-middleware. Frontend port 8080, backend port 3001.
Backend Deps: spring-boot-starter-web, spring-boot-starter-jdbc, sqlite-jdbc 3.46.1.3, spring-boot-starter-test.
Frontend Deps: axios ^1.4.0, chart.js ^4.5.1, react ^18.2.0, react-chartjs-2 ^5.3.1, react-dom ^18.2.0, react-router-dom ^6.13.0.


Project Structure — Backend:
backend-api/src/main/java/com/svpoc/backendjava/
  BackendJavaApplication.java
  config/ -> CorsConfig, DataSourceConfig, ProxyFilter
  controller/ -> ApiRegistryController, AuthController, HealthCheckController, HealthController, ProxyController, RequestController, SettingsController
  model/ -> ApiRegistryEntry, Request, Setting, Stub, StubVersion, User
  service/ -> ApiRegistryService, CoreService, DbService, HttpForwardService, JwtService, UserService
  stub/controller/ -> StubController, StubVersionController
  stub/dto/ -> StubRequest
  stub/service/ -> StubMatchingService, StubService
  util/ -> ConversionUtil, JsonUtil, SecurityHelper
Frontend: frontend-ui/src/
  App.js, index.js, setupProxy.js, services/registryApi.js
  components/analytics/ -> AnalyticsTab
  components/auth/ -> LoginPage, RegisterPage


  Frontend continued:
  common/ -> Dashboard, Header, Sidebar, Footer, registryStyles
  health/ -> ApiHealth, EndpointCard, healthCheckRequests, useEndpointHealthCheck
  registry/ -> ApiFormModal, ApiRegistry
  requests/ -> RequestDetail, RequestsTab
  settings/ -> SettingsTab; stubs/ -> StubForm, StubsTab; tester/ -> APITesterTab
RBAC Groups:
1. admin (Admin): Full CRUD, clear data, manage users
2. QED_DEV_OPS (Dev Ops): Read+Write (stubs, requests, settings)
3. QED_DEFAULT_USER (Default User): Read-only
Auth Flow:
1. Register (POST /api/auth/register): username, SHA-256 password, optional email, AD group (QED_DEV_OPS/QED_DEFAULT_USER for self-reg). Returns JWT.
2. Login (POST /api/auth/login): username + hashed password → JWT {sub, role, adGroup, iat, exp}.
3. Verify (GET /api/auth/verify-token): validates Bearer, returns claims.
4. Logout (POST /api/auth/logout): client clears localStorage.


JWT Implementation:
- Custom HMAC-SHA256 signing (no Spring Security OAuth)
- Secret from env var JWT_SECRET (fallback to dev default)
- Expiration: 24 hours (jwt.expiration=86400000)
- Token: { sub, role, adGroup, iat, exp }
- Protected endpoints use SecurityHelper.extractUser(authorization, jwtService)
Password Security:
- Client: SHA-256 hash before transmission (Web Crypto API)
- Server: PBKDF2WithHmacSHA256 (65536 iterations, 256-bit key) with 16-byte salt
- Stored as base64(salt):base64(hash)
Default Admin: seeded on startup if SV_ADMIN_PASSWORD env var is set. Auto-fixes role/adGroup.
SecurityHelper:
- extractUser() — user info from Bearer token
- isAdmin() — adGroup is "admin"
- canWrite() — admin or QED_DEV_OPS
- canDelete() — admin only
- resolveToggleValue() — normalizes to "on"/"off"
- firstNonBlank() — returns first non-blank string


Feature 1: Health Monitoring
Sub-module 1A: API Registry (/api-registry)
Backend: ApiRegistryController + ApiRegistryService
Table api_registry:
CREATE TABLE IF NOT EXISTS api_registry (
  id TEXT PRIMARY KEY, functionName TEXT, method TEXT, endpoint TEXT,
  category TEXT, environment TEXT, description TEXT, healthCheckUrl TEXT,
  healthCheckHeaders TEXT, healthCheckBody TEXT, healthCheckParams TEXT,
  retryOn500 INTEGER DEFAULT 0, isCustom BOOLEAN DEFAULT 0,
  ownerGroup TEXT DEFAULT 'admin', createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
)
Endpoints:
- GET /api/registry — list all (filtered by env/category, ownerGroup scoped for non-admin)
- GET /api/registry/{id} — get by ID (ownership check)
- POST /api/registry — create (write perm)
- PATCH /api/registry/{id} — update (ownership check)
- DELETE /api/registry/{id} — delete (admin)
- DELETE /api/registry — delete all (admin)



Frontend: ApiRegistry.js + ApiFormModal.js
- Table: method badge, name, endpoint, category, actions
- "Register API" opens modal; supports cURL import
- Form: API Name, Method, Category, URL, Environment, Description, Headers, Body
- Edit/Delete (role-gated)
- registryApi.js: axios with JWT Bearer token
- CRUD: getAllApis(), createApi(), updateApi(), deleteApi()
Sub-module 1B: API Health Dashboard (/api-health)
Backend: HealthCheckController
- POST /api/health-check — server-side proxy (avoids CORS)
- Accepts { url, method, headers, body, params }
- Returns { statusCode, responseTime, body, headers } or { error }
- 15s timeout, uses HttpForwardService.send()
- Requires auth
Frontend: ApiHealth.js + EndpointCard.js
- Dashboard: registered APIs as expandable cards
- Fetches on mount, polls every 5 minutes


ApiHealth.js continued:
- Filters: status pills (ALL/UP/DOWN), environment, category, text search
- Summary bar: filtered vs total count
- Edit/Delete on cards (role-gated)
- "Refresh All" re-fetches registry + re-runs checks
- Detects new endpoints by comparing function name hashes
EndpointCard.js:
- Collapsed: method badge, name, description, category, status dot, expand arrow
- Expanded: URL bar, headers, params + body, status chips (code, time, last checked, env), error
useEndpointHealthCheck.js (custom hook):
- Module-level cache (survives unmount/remount)
- Progressive updates: checks each endpoint, updates state as each resolves
- 30-second polling interval
- checkNewEndpoints() for newly registered APIs
- refreshEndpoints() for manual refresh
- Returns: { endpointStatusMap, isChecking, lastChecked, refreshEndpoints, checkNewEndpoints }


healthCheckRequests.js:
- checkApiHealth(api) — single endpoint via POST /api/health-check
- Retry on 500 up to retryOn500 times, 2s delays; also retries null statusCode
- Returns: { name, status, statusCode, responseTime, lastChecked, error, retryCount }
- checkEndpointListHealth() — batch with dependency chain support
- Dependency chain: postAccount runs first, feeds into createSubmissions
Feature 2: Service Virtualization
Sub-module 2A: API Tester (/service-virtualization/api)
Backend: ProxyController
1. POST /api/proxy-external — full URLs (http/https)
   - SSRF protection: blocks internal/private/loopback/site-local addresses
   - Toggle from headers (X-Use-Toggle/X-Mock-Toggle) or body useStubs
   - OFF: returns matching stub (with delay)
   - ON: forwards upstream, records, auto-creates stub on 2xx


 proxy-external continued:
   - Records every request regardless of outcome
   - Returns: { success, status, data, source, recorded, recordedId }
2. POST /api/proxy-request — path-only endpoints
   - Target host from targetHost or X-Target-Host header
   - Constructs full URL from target host + endpoint path
   - OFF: stub match; multiple hosts → { availableHosts }; single → stub
   - ON: forwards upstream, records, creates stub
     - Failure: falls back to stub, else 503
   - Returns: { status, data, headers, source, recordedId, fallback? }
ProxyFilter (Servlet Filter — catch-all):
- Catches non-/api/* and non-/health paths
- Requires Bearer token (401 if missing)
- Toggle from X-Use-Toggle/X-Mock-Toggle or settings DB
- Target host from X-Target-Host header
- Same stub matching + recording logic as ProxyController 

ProxyFilter continued:
- ON + target host: forwards upstream, records + creates stub on 2xx-3xx
- Upstream error (4xx-5xx): falls back to stub
- Upstream exception: falls back to stub, or 503
ProxyFilter header handling:
- Request sanitization: strips host, connection, content-length, transfer-encoding, accept-encoding, postman-token, x-use-toggle, x-mock-toggle, x-target-host, x-api-category
- Response sanitization: strips transfer-encoding, connection, keep-alive, proxy-auth headers, te, trailers, upgrade, content-encoding, :status
- Validates RFC-compliant header names, strips CR/LF from values
- Writes X-Response-Source header ("stub" or "live-api")
HttpForwardService:
- java.net.http.HttpClient, 5-second connect timeout
- Handles gzip/deflate decompression manually
- Falls back to raw UTF-8 if decompression fails
- OutboundResponse: (status, body, headers)

Frontend: APITesterTab.js
Form fields:
- Base URL (upstream domain, sent as X-Target-Host)
- HTTP Method select (GET/POST/PUT/DELETE/PATCH)
- API Category select (Claims/SBI/GW/Other)
- Endpoint Path input (path-only or full URL)
- Request Body textarea (JSON, non-GET only)
- Request Headers textarea (JSON)
- Toggle indicator badge (ON/OFF)
Header sanitization: strips hop-by-hop headers (host, connection, keep-alive, transfer-encoding, content-length, accept-encoding, postman-token, x-use-toggle, x-mock-toggle, x-target-host), unwraps double-quoted values.
Routing: http/https URL → /api/proxy-external; path-only → /api/proxy-request.
Injects X-Target-Host if Base URL provided.
15-second timeout, validateStatus accepts all codes.
Response panel:
- Status badge (2xx green, 3xx blue, 4xx yellow, 5xx red)
- Response time, source badge (Live API / From Stub)

APITesterTab response panel continued:
- Response data as formatted JSON
- Response headers as formatted JSON
- Copy response to clipboard
- Error box for network errors
- Read-only mode for QED_DEFAULT_USER (disabled send + lock icon)
- Supports initialData prop from other tabs
Sub-module 2B: Mock Stubs (/service-virtualization/stubs)
Backend: StubController + StubVersionController + StubService + StubMatchingService
Database table stubs:
CREATE TABLE IF NOT EXISTS stubs (
  id TEXT PRIMARY KEY, name TEXT NOT NULL, method TEXT NOT NULL,
  urlPattern TEXT NOT NULL, targetHost TEXT DEFAULT '',
  requestMatcher TEXT, responseStatus INTEGER DEFAULT 200,
  responseBody TEXT, responseHeaders TEXT,
  delay INTEGER DEFAULT 0, enabled BOOLEAN DEFAULT 1,
  category TEXT DEFAULT 'other', version TEXT DEFAULT 'v1',
  ownerGroup TEXT DEFAULT 'admin',
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP


  Table stub_versions:
CREATE TABLE IF NOT EXISTS stub_versions (
  versionId TEXT PRIMARY KEY, stubId TEXT NOT NULL,
  version TEXT NOT NULL, versionTag TEXT,
  responseStatus INTEGER DEFAULT 200, responseBody TEXT,
  responseHeaders TEXT, isActive BOOLEAN DEFAULT 0,
  createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (stubId) REFERENCES stubs(id) ON DELETE CASCADE
)
Stub CRUD:
- GET /api/stubs — list all
- POST /api/stubs — create/update (upsert by method+urlPattern+targetHost+category)
- PUT /api/stubs/{id} — update
- DELETE /api/stubs/{id} — delete + cascade versions
- DELETE /api/stubs — clear all
- POST /api/stubs/{id}/toggle — toggle enabled
Version endpoints:
- GET /api/stubs/{stubId}/versions — list (ordered by version, status)
- GET /api/stubs/{stubId}/versions/{versionId} — detail


Version endpoints continued:
- POST /api/stubs/{stubId}/versions — create version:
  - Rejects if same responseStatus exists
  - Auto-assigns next version number (finds gaps)
  - Auto-activates if status 200 or first version
- PUT /api/stubs/{stubId}/versions/{versionId} — update (rejects duplicate status)
- POST /api/stubs/{stubId}/versions/{versionId}/activate:
  - Deactivates all other versions
  - Syncs master stub responseStatus/responseBody/responseHeaders
- DELETE /api/stubs/{stubId}/versions/{versionId} — cannot delete active version
Version Strategy:
- Maps to HTTP status ranges: v1=2xx, v2=4xx, v3=3xx, v4=5xx
- Master stub per unique (urlPattern + targetHost + category)
- Each status code gets its own version entry
- Only one version active at a time
- Auto-recording: existing versions updated, new statuses get inactive versions


Stub Matching Algorithm (CoreService):
1. Query enabled stubs for the HTTP method
2. URL match: ^ or .* → regex; otherwise substring
3. Body matcher: parse requestMatcher JSON, verify key-values
4. Priority: exact host > host-agnostic > category > other
5. No targetHost + multiple hosts → returns availableHosts
Frontend: StubsTab.js + StubForm.js
StubsTab.js:
- Left panel: stub list — method badge, URL pattern, host, enabled badge, version count
- Category filter (All/Claims/SBI/GW/Other)
- Controls: Create, Refresh, Clear All (admin), count
- Auto-refresh every 5 seconds
- Fetches versions per stub in parallel
Right panel (stub selected):
- Name, category, method, URL pattern, target host
- Version list with active badge, status, Activate/Delete


StubsTab right panel continued:
- Actions: Create Version, Toggle, Edit, Delete, Test in API Tester
Version detail (version selected):
- Back button, version badge, active indicator
- Status (color-coded), headers, body (code blocks)
- Actions: Edit, Activate (if inactive), Delete (if inactive)
VersionFormModal:
- Fields: Status Code, Headers (JSON), Body
- For create and edit modes
StubForm.js (modal):
- Fields: Name, Method, URL pattern, Category, Target Host, Status, Delay, Body, Headers, Request Matcher
- Edit: pre-fills; Create: empty
- POST /api/stubs or PUT /api/stubs/{id}
Export/Import:
- Export: JSON file (exportedAt, version, stubs array)
- Import: reads JSON, POSTs each stub (upsert)


Sub-module 2C: Recorded Requests (/service-virtualization/requests)
Backend: RequestController
Database table requests:
CREATE TABLE IF NOT EXISTS requests (
  id TEXT PRIMARY KEY, method TEXT NOT NULL, url TEXT NOT NULL,
  baseUrl TEXT, endpoint TEXT, headers TEXT NOT NULL, body TEXT,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP, status INTEGER,
  response TEXT, responseHeaders TEXT, isRecorded BOOLEAN DEFAULT 0,
  category TEXT DEFAULT 'other', ownerGroup TEXT DEFAULT 'admin'
)
Endpoints:
- GET /api/recorded-requests?limit=100 — list (scoped by ownerGroup)
- GET /api/requests?useToggle=on|off&limit=100 — off→stubs, on→requests
- GET /api/requests/{id} — single request (ownership check)
- POST /api/record-test — manually record (write perm, auto-derives endpoint/baseUrl)
- DELETE /api/requests — clear all (admin only)
- DELETE /api/requests/{id} — delete single (admin only)


Frontend: RequestsTab.js + RequestDetail.js
RequestsTab.js:
- Left panel: request list
  - Each: method badge, URL, source badge (Live=green/Stub=blue), category badge, status badge, timestamp
  - Base URL below with globe icon
- Controls: text filter, category dropdown, Refresh, Clear All (admin), count
- Auto-refresh every 5 minutes
- forwardRef exposing refreshRequests()
RequestDetail.js:
- Header: method + URL, source badge, Test in API Tester button, close
- Sections: Status, Timestamp, Base URL, Data Source, Category
- Code blocks: Request Headers, Request Body, Response Headers, Response Body
- Delete button (admin only)
Sub-module 2D: Analytics (/service-virtualization/analytics)
Frontend: AnalyticsTab.js
- Fetches from GET /api/recorded-requests?limit=1000
- API search input + dropdown to filter by endpoint
- Auto-selects single match



Charts (Chart.js):
1. API Call Frequency — Line chart by hour
2. Response Time Trends — Line: avg (filled), min/max (dashed)
3. Error Rate — Pie (2xx vs 4xx/5xx), stat boxes
4. Most Used Endpoints — ranked list, count, progress bar
Summary: total requests, avg response time, unique endpoints, success rate.
Sub-module 2E: Settings (/service-virtualization/settings)
Backend: SettingsController
Table: settings (key TEXT PRIMARY KEY, value TEXT)
Defaults: recordingMode=true, playbackMode=true, targetUrl=(empty), useToggle=true
Endpoints:
- GET /api/settings — all as map (auth required)
- PUT /api/settings/{key} — update (write perm)
- Keys: recordingMode, playbackMode, targetUrl, useToggle, dashboard_username, dashboard_password, session_*
Frontend: SettingsTab.js
- Server Config, Operation Modes, How It Works, API Endpoints cards


Sidebar (Sidebar.js) — Two nav groups:
1. 🩺 Health Monitoring: 📚 API Registry (/api-registry), 💓 API Health (/api-health)
2. 🧪 Service Virtualization: 🔌 API Request, 🎭 Stubs, 📋 Recorded Requests, 📊 Analytics, ⚙️ Settings (under /service-virtualization/)
Inline: Real Time API toggle (ON=live, OFF=stubs), Recording Mode toggle
Core Services:
CoreService:
- recordRequest(Map) — INSERT OR REPLACE into requests
- findMatchingStub(method, url, body, targetHost) — host/category priority
- findMatchingStubWithHostOptions() — returns stub + availableHosts
- createOrReplaceStub() — upserts master stub + versions
- parseRequestRow() / parseStubRow() — JSON field parsing
- determineVersion(statusCode) — 2xx→v1, 4xx→v2, 3xx→v3, 5xx→v4



CoreService helpers: stringVal(), intVal(), boolVal(), normalizeHost(), matchesUrlPattern(), matchesRequestMatcher()
DbService:
- Wraps JdbcTemplate: query(), querySingle(), update(), getSetting(), upsertSetting(), getAllSettings()
- initializeDatabase() — creates tables (requests, stubs, stub_versions, settings, users, api_registry) with ALTER TABLE migrations wrapped in try/catch
- Backfills missing stub_versions for orphan stubs on startup
HttpForwardService:
- java.net.http.HttpClient, 5s connect timeout
- Handles gzip/deflate decompression, strips Accept-Encoding
- OutboundResponse record: (status, body, headers)
- singleValueHeaders() — multi-value to single-value map
DataSourceConfig:
- SQLite path: DB_PATH env > ./backend/data.db > ../backend/data.db > ./data/data.db
- Creates parent dirs, HikariCP maximumPoolSize=1


Config — application.properties:
  server.port=3001, jwt.secret=${JWT_SECRET:change-this-secret}
  jwt.expiration=86400000, admin.default-password=${SV_ADMIN_PASSWORD:}
application.yml: port 3001, error.include-message=always, app sv-poc-backend (v1.0.0), jackson non_null, actuator health+info+metrics
CORS: localhost:3000/:3001/:8080 + 127.0.0.1 equivalents
setupProxy.js: proxies /api and /v1 to localhost:3001
Key Notes:
1. Toggle: headers > body > DB settings
2. Auto-Stub: live 2xx auto-saved as stubs; if upstream goes down, proxy falls back to recorded stub
3. Multi-Host: scoped by (urlPattern + targetHost + category)
4. Data Isolation: ownerGroup tagging, non-admin sees own + admin
5. SSRF: blocks internal/private/loopback
6. Header Sanitization: strips hop-by-hop + tool headers
7. No Docker: SQLite file DB, Java 21 + Node.js only
8. Progressive Health: cards update individually