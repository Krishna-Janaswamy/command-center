package virtualization

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) init() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS virtualization_apis (
			id TEXT PRIMARY KEY, name TEXT, version TEXT DEFAULT 'v1', method TEXT, endpoint TEXT, base_url TEXT,
			environment TEXT, category TEXT, description TEXT, health_check_headers TEXT, health_check_body TEXT,
			health_check_params TEXT, retry_on_500 INTEGER DEFAULT 0, is_custom INTEGER DEFAULT 0,
			owner_group TEXT DEFAULT 'admin', enabled INTEGER DEFAULT 1, created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS virtualization_stubs (
			id TEXT PRIMARY KEY, api_id TEXT, name TEXT, method TEXT, endpoint TEXT, base_url TEXT,
			environment TEXT, category TEXT, description TEXT, owner_group TEXT, request_matcher TEXT,
			response_status INTEGER DEFAULT 200, response_body TEXT, response_headers TEXT, delay_ms INTEGER DEFAULT 0,
			enabled INTEGER DEFAULT 1, version TEXT DEFAULT 'v1', created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS virtualization_versions (
			id TEXT PRIMARY KEY, stub_id TEXT, version TEXT, version_tag TEXT, response_status INTEGER DEFAULT 200,
			response_body TEXT, response_headers TEXT, active INTEGER DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS virtualization_requests (
			id TEXT PRIMARY KEY, method TEXT, url TEXT, base_url TEXT, endpoint TEXT, headers TEXT, body TEXT,
			body_s3_key TEXT, status INTEGER, response TEXT, response_s3_key TEXT, response_headers TEXT,
			is_recorded INTEGER DEFAULT 0, category TEXT, owner_group TEXT, source TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS virtualization_settings (key TEXT PRIMARY KEY, value TEXT);`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) ListAPIs() []*API {
	rows, err := s.db.Query("SELECT id, name, version, method, endpoint, base_url, environment, category, description, health_check_headers, health_check_body, health_check_params, retry_on_500, is_custom, owner_group, enabled, created_at FROM virtualization_apis ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*API{}
	for rows.Next() {
		v := &API{}
		var enabledInt, customInt int
		if err := rows.Scan(&v.ID, &v.Name, &v.Version, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.HealthCheckHeaders, &v.HealthCheckBody, &v.HealthCheckParams, &v.RetryOn500, &customInt, &v.OwnerGroup, &enabledInt, &v.CreatedAt); err == nil {
			v.Enabled = enabledInt == 1
			v.IsCustom = customInt == 1
			out = append(out, v)
		}
	}
	return out
}

func (s *SQLiteStore) GetAPI(id string) (*API, bool) {
	v := &API{}
	var enabledInt, customInt int
	err := s.db.QueryRow("SELECT id, name, version, method, endpoint, base_url, environment, category, description, health_check_headers, health_check_body, health_check_params, retry_on_500, is_custom, owner_group, enabled, created_at FROM virtualization_apis WHERE id=?", id).Scan(&v.ID, &v.Name, &v.Version, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.HealthCheckHeaders, &v.HealthCheckBody, &v.HealthCheckParams, &v.RetryOn500, &customInt, &v.OwnerGroup, &enabledInt, &v.CreatedAt)
	if err != nil {
		return nil, false
	}
	v.Enabled = enabledInt == 1
	v.IsCustom = customInt == 1
	return v, true
}

func (s *SQLiteStore) SaveAPI(v *API) *API {
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	if !v.Enabled {
		v.Enabled = true
	}
	enabledInt := 0
	if v.Enabled {
		enabledInt = 1
	}
	customInt := 0
	if v.IsCustom {
		customInt = 1
	}
	_, _ = s.db.Exec(`INSERT INTO virtualization_apis (id, name, version, method, endpoint, base_url, environment, category, description, health_check_headers, health_check_body, health_check_params, retry_on_500, is_custom, owner_group, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, version=excluded.version, method=excluded.method, endpoint=excluded.endpoint, base_url=excluded.base_url, environment=excluded.environment, category=excluded.category, description=excluded.description, health_check_headers=excluded.health_check_headers, health_check_body=excluded.health_check_body, health_check_params=excluded.health_check_params, retry_on_500=excluded.retry_on_500, is_custom=excluded.is_custom, owner_group=excluded.owner_group, enabled=excluded.enabled`,
		v.ID, v.Name, v.Version, v.Method, v.Endpoint, v.BaseURL, v.Environment, v.Category, v.Description, v.HealthCheckHeaders, v.HealthCheckBody, v.HealthCheckParams, v.RetryOn500, customInt, v.OwnerGroup, enabledInt, v.CreatedAt)
	result, _ := s.GetAPI(v.ID)
	return result
}

func (s *SQLiteStore) DeleteAPI(id string) bool {
	res, err := s.db.Exec("DELETE FROM virtualization_apis WHERE id=?", id)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *SQLiteStore) ListStubs() []*Stub {
	rows, err := s.db.Query("SELECT id, api_id, name, method, endpoint, base_url, environment, category, description, owner_group, request_matcher, response_status, response_body, response_headers, delay_ms, enabled, version FROM virtualization_stubs ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*Stub{}
	for rows.Next() {
		v := &Stub{}
		var enabledInt int
		if err := rows.Scan(&v.ID, &v.APIID, &v.Name, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.OwnerGroup, &v.RequestMatcher, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.DelayMS, &enabledInt, &v.Version); err == nil {
			v.Enabled = enabledInt == 1
			out = append(out, v)
		}
	}
	return out
}

func (s *SQLiteStore) GetStub(id string) (*Stub, bool) {
	v := &Stub{}
	var enabledInt int
	err := s.db.QueryRow("SELECT id, api_id, name, method, endpoint, base_url, environment, category, description, owner_group, request_matcher, response_status, response_body, response_headers, delay_ms, enabled, version FROM virtualization_stubs WHERE id=?", id).Scan(&v.ID, &v.APIID, &v.Name, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.OwnerGroup, &v.RequestMatcher, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.DelayMS, &enabledInt, &v.Version)
	if err != nil {
		return nil, false
	}
	v.Enabled = enabledInt == 1
	return v, true
}

func (s *SQLiteStore) SaveStub(v *Stub) *Stub {
	isNew := v.ID == ""
	if isNew {
		v.ID = newID()
	}
	if v.Method == "" {
		v.Method = "ANY"
	}
	if v.Version == "" {
		v.Version = "v1"
	}
	if v.Environment == "" {
		v.Environment = "Dev"
	}
	enabledInt := 0
	if v.Enabled {
		enabledInt = 1
	}
	_, _ = s.db.Exec(`INSERT INTO virtualization_stubs (id, api_id, name, method, endpoint, base_url, environment, category, description, owner_group, request_matcher, response_status, response_body, response_headers, delay_ms, enabled, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET api_id=excluded.api_id, name=excluded.name, method=excluded.method, endpoint=excluded.endpoint, base_url=excluded.base_url, environment=excluded.environment, category=excluded.category, description=excluded.description, owner_group=excluded.owner_group, request_matcher=excluded.request_matcher, response_status=excluded.response_status, response_body=excluded.response_body, response_headers=excluded.response_headers, delay_ms=excluded.delay_ms, enabled=excluded.enabled, version=excluded.version`,
		v.ID, v.APIID, v.Name, v.Method, v.Endpoint, v.BaseURL, v.Environment, v.Category, v.Description, v.OwnerGroup, v.RequestMatcher, v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, v.DelayMS, enabledInt, v.Version)

	if isNew {
		v1 := &StubVersion{
			ID:              newID(),
			StubID:          v.ID,
			Version:         v.Version,
			VersionTag:      "Initial version",
			ResponseStatus:  v.ResponseStatus,
			ResponseBody:    v.ResponseBody,
			ResponseHeaders: v.ResponseHeaders,
			Active:          true,
		}
		s.SaveVersion(v1)
	}

	result, _ := s.GetStub(v.ID)
	return result
}

func (s *SQLiteStore) DeleteStub(id string) bool {
	_, _ = s.db.Exec("DELETE FROM virtualization_versions WHERE stub_id=?", id)
	res, err := s.db.Exec("DELETE FROM virtualization_stubs WHERE id=?", id)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *SQLiteStore) ClearStubs() {
	_, _ = s.db.Exec("DELETE FROM virtualization_versions")
	_, _ = s.db.Exec("DELETE FROM virtualization_stubs")
}

func (s *SQLiteStore) ToggleStub(id string) (*Stub, bool) {
	_, err := s.db.Exec("UPDATE virtualization_stubs SET enabled = CASE WHEN enabled = 1 THEN 0 ELSE 1 END WHERE id=?", id)
	if err != nil {
		return nil, false
	}
	return s.GetStub(id)
}

func (s *SQLiteStore) FindStub(method, url, body string) (*Stub, bool) {
	return s.FindStubForAPI("", method, url, body)
}

func (s *SQLiteStore) FindStubForAPI(apiID, method, url, body string) (*Stub, bool) {
	for _, v := range s.ListStubs() {
		if v.Enabled && (apiID == "" || v.APIID == "" || v.APIID == apiID) && methodMatches(v.Method, method) && pathMatches(v.Endpoint, url) && bodyMatches(v.RequestMatcher, body) {
			return v, true
		}
	}
	return nil, false
}

func (s *SQLiteStore) ListVersions(stubID string) []*StubVersion {
	rows, err := s.db.Query("SELECT id, stub_id, version, version_tag, response_status, response_body, response_headers, active FROM virtualization_versions WHERE stub_id=?", stubID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*StubVersion{}
	for rows.Next() {
		v := &StubVersion{}
		var activeInt int
		if err := rows.Scan(&v.ID, &v.StubID, &v.Version, &v.VersionTag, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &activeInt); err == nil {
			v.Active = activeInt == 1
			out = append(out, v)
		}
	}
	return out
}

func (s *SQLiteStore) SaveVersion(v *StubVersion) *StubVersion {
	if v.ID == "" {
		v.ID = newID()
	}
	activeInt := 0
	if v.Active {
		activeInt = 1
	}
	_, _ = s.db.Exec(`INSERT INTO virtualization_versions (id, stub_id, version, version_tag, response_status, response_body, response_headers, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET version=excluded.version, version_tag=excluded.version_tag, response_status=excluded.response_status, response_body=excluded.response_body, response_headers=excluded.response_headers, active=excluded.active`,
		v.ID, v.StubID, v.Version, v.VersionTag, v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, activeInt)

	if v.Active {
		_, _ = s.db.Exec("UPDATE virtualization_stubs SET response_status=?, response_body=?, response_headers=? WHERE id=?", v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, v.StubID)
	}
	return v
}

func (s *SQLiteStore) DeleteVersion(id string) bool {
	res, err := s.db.Exec("DELETE FROM virtualization_versions WHERE id=?", id)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *SQLiteStore) ActivateVersion(stubID, versionID string) bool {
	tx, err := s.db.Begin()
	if err != nil {
		return false
	}
	defer tx.Rollback()

	if _, err = tx.Exec("UPDATE virtualization_versions SET active=0 WHERE stub_id=?", stubID); err != nil {
		return false
	}
	if _, err = tx.Exec("UPDATE virtualization_versions SET active=1 WHERE id=? AND stub_id=?", versionID, stubID); err != nil {
		return false
	}
	v, ok := s.GetVersion(versionID)
	if !ok {
		return false
	}
	if _, err = tx.Exec("UPDATE virtualization_stubs SET response_status=?, response_body=?, response_headers=? WHERE id=?", v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, stubID); err != nil {
		return false
	}
	return tx.Commit() == nil
}

func (s *SQLiteStore) GetVersion(id string) (*StubVersion, bool) {
	v := &StubVersion{}
	var activeInt int
	err := s.db.QueryRow("SELECT id, stub_id, version, version_tag, response_status, response_body, response_headers, active FROM virtualization_versions WHERE id=?", id).Scan(&v.ID, &v.StubID, &v.Version, &v.VersionTag, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &activeInt)
	if err != nil {
		return nil, false
	}
	v.Active = activeInt == 1
	return v, true
}

func (s *SQLiteStore) SaveRequest(v *RequestRecord) *RequestRecord {
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	recInt := 0
	if v.IsRecorded {
		recInt = 1
	}
	_, _ = s.db.Exec(`INSERT INTO virtualization_requests (id, method, url, base_url, endpoint, headers, body, body_s3_key, status, response, response_s3_key, response_headers, is_recorded, category, owner_group, source, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.ID, v.Method, v.URL, v.BaseURL, v.Endpoint, v.Headers, v.Body, v.BodyS3Key, v.Status, v.Response, v.ResponseS3Key, v.ResponseHeaders, recInt, v.Category, v.OwnerGroup, v.Source, v.CreatedAt)
	return v
}

func (s *SQLiteStore) ListRequests() []*RequestRecord {
	rows, err := s.db.Query("SELECT id, method, url, base_url, endpoint, headers, body, body_s3_key, status, response, response_s3_key, response_headers, is_recorded, category, owner_group, source, created_at FROM virtualization_requests ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*RequestRecord{}
	for rows.Next() {
		v := &RequestRecord{}
		var recInt int
		if err := rows.Scan(&v.ID, &v.Method, &v.URL, &v.BaseURL, &v.Endpoint, &v.Headers, &v.Body, &v.BodyS3Key, &v.Status, &v.Response, &v.ResponseS3Key, &v.ResponseHeaders, &recInt, &v.Category, &v.OwnerGroup, &v.Source, &v.CreatedAt); err == nil {
			v.IsRecorded = recInt == 1
			out = append(out, v)
		}
	}
	return out
}

func (s *SQLiteStore) DeleteRequest(id string) bool {
	res, err := s.db.Exec("DELETE FROM virtualization_requests WHERE id=?", id)
	if err != nil {
		return false
	}
	affected, _ := res.RowsAffected()
	return affected > 0
}

func (s *SQLiteStore) ClearRequests() {
	_, _ = s.db.Exec("DELETE FROM virtualization_requests")
}

func (s *SQLiteStore) SetSetting(k, v string) {
	_, _ = s.db.Exec("INSERT INTO virtualization_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", k, v)
}

func (s *SQLiteStore) Setting(k string) string {
	var v string
	_ = s.db.QueryRow("SELECT value FROM virtualization_settings WHERE key=?", k).Scan(&v)
	return v
}
