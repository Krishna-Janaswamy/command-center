package virtualization

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewConfiguredStore(ctx context.Context) Store {
	url := os.Getenv("DATABASE_URL")
	if url != "" {
		pool, err := pgxpool.New(ctx, url)
		if err == nil {
			store := &PostgresStore{pool: pool}
			if err := store.init(ctx); err == nil {
				return store
			}
			pool.Close()
		}
	}
	sqlitePath := os.Getenv("SQLITE_PATH")
	if sqlitePath == "" {
		sqlitePath = "../backend/data/data.db"
	}
	if sqliteStore, err := NewSQLiteStore(sqlitePath); err == nil {
		return sqliteStore
	}
	return NewMemoryStore()
}

func (s *PostgresStore) init(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS virtualization_apis (id TEXT PRIMARY KEY, name TEXT, version TEXT DEFAULT 'v1', method TEXT, endpoint TEXT, base_url TEXT, environment TEXT, category TEXT, description TEXT, health_check_headers TEXT, health_check_body TEXT, health_check_params TEXT, retry_on_500 INTEGER DEFAULT 0, is_custom BOOLEAN DEFAULT FALSE, owner_group TEXT DEFAULT 'admin', enabled BOOLEAN DEFAULT TRUE, created_at TIMESTAMPTZ DEFAULT NOW());
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS version TEXT;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS health_check_headers TEXT;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS health_check_body TEXT;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS health_check_params TEXT;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS retry_on_500 INTEGER;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS is_custom BOOLEAN;
ALTER TABLE virtualization_apis ADD COLUMN IF NOT EXISTS owner_group TEXT;
CREATE TABLE IF NOT EXISTS virtualization_stubs (id TEXT PRIMARY KEY, api_id TEXT, name TEXT, method TEXT, endpoint TEXT, base_url TEXT, environment TEXT, category TEXT, description TEXT, owner_group TEXT, request_matcher TEXT, response_status INTEGER, response_body TEXT, response_headers TEXT, delay_ms INTEGER, enabled BOOLEAN DEFAULT TRUE, version TEXT, created_at TIMESTAMPTZ DEFAULT NOW());
ALTER TABLE virtualization_stubs ADD COLUMN IF NOT EXISTS environment TEXT;
ALTER TABLE virtualization_stubs ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE virtualization_stubs ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE virtualization_stubs ADD COLUMN IF NOT EXISTS owner_group TEXT;
CREATE TABLE IF NOT EXISTS virtualization_requests (id TEXT PRIMARY KEY, method TEXT, url TEXT, base_url TEXT, endpoint TEXT, headers TEXT, body TEXT, body_s3_key TEXT, status INTEGER, response TEXT, response_s3_key TEXT, response_headers TEXT, is_recorded BOOLEAN DEFAULT FALSE, category TEXT, owner_group TEXT, source TEXT, created_at TIMESTAMPTZ DEFAULT NOW());
ALTER TABLE virtualization_requests ADD COLUMN IF NOT EXISTS base_url TEXT;
ALTER TABLE virtualization_requests ADD COLUMN IF NOT EXISTS is_recorded BOOLEAN;
ALTER TABLE virtualization_requests ADD COLUMN IF NOT EXISTS category TEXT;
ALTER TABLE virtualization_requests ADD COLUMN IF NOT EXISTS owner_group TEXT;
CREATE TABLE IF NOT EXISTS virtualization_versions (id TEXT PRIMARY KEY, stub_id TEXT, version TEXT, version_tag TEXT, response_status INTEGER, response_body TEXT, response_headers TEXT, active BOOLEAN DEFAULT FALSE);
ALTER TABLE virtualization_versions ADD COLUMN IF NOT EXISTS version_tag TEXT;
CREATE TABLE IF NOT EXISTS virtualization_settings (key TEXT PRIMARY KEY, value TEXT);`)
	return err
}
func (s *PostgresStore) ListAPIs() []*API {
	rows, err := s.pool.Query(context.Background(), "SELECT id,name,version,method,endpoint,base_url,environment,category,description,health_check_headers,health_check_body,health_check_params,retry_on_500,is_custom,owner_group,enabled,created_at FROM virtualization_apis ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*API{}
	for rows.Next() {
		v := &API{}
		if rows.Scan(&v.ID, &v.Name, &v.Version, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.HealthCheckHeaders, &v.HealthCheckBody, &v.HealthCheckParams, &v.RetryOn500, &v.IsCustom, &v.OwnerGroup, &v.Enabled, &v.CreatedAt) == nil {
			out = append(out, v)
		}
	}
	return out
}
func (s *PostgresStore) GetAPI(id string) (*API, bool) {
	v := &API{}
	err := s.pool.QueryRow(context.Background(), "SELECT id,name,version,method,endpoint,base_url,environment,category,description,health_check_headers,health_check_body,health_check_params,retry_on_500,is_custom,owner_group,enabled,created_at FROM virtualization_apis WHERE id=$1", id).Scan(&v.ID, &v.Name, &v.Version, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.HealthCheckHeaders, &v.HealthCheckBody, &v.HealthCheckParams, &v.RetryOn500, &v.IsCustom, &v.OwnerGroup, &v.Enabled, &v.CreatedAt)
	return v, err == nil
}
func (s *PostgresStore) SaveAPI(v *API) *API {
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	if !v.Enabled {
		v.Enabled = true
	}
	_, _ = s.pool.Exec(context.Background(), `INSERT INTO virtualization_apis (id,name,version,method,endpoint,base_url,environment,category,description,health_check_headers,health_check_body,health_check_params,retry_on_500,is_custom,owner_group,enabled,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) ON CONFLICT(id) DO UPDATE SET name=EXCLUDED.name,version=EXCLUDED.version,method=EXCLUDED.method,endpoint=EXCLUDED.endpoint,base_url=EXCLUDED.base_url,environment=EXCLUDED.environment,category=EXCLUDED.category,description=EXCLUDED.description,health_check_headers=EXCLUDED.health_check_headers,health_check_body=EXCLUDED.health_check_body,health_check_params=EXCLUDED.health_check_params,retry_on_500=EXCLUDED.retry_on_500,is_custom=EXCLUDED.is_custom,owner_group=EXCLUDED.owner_group,enabled=EXCLUDED.enabled`, v.ID, v.Name, v.Version, v.Method, v.Endpoint, v.BaseURL, v.Environment, v.Category, v.Description, v.HealthCheckHeaders, v.HealthCheckBody, v.HealthCheckParams, v.RetryOn500, v.IsCustom, v.OwnerGroup, v.Enabled, v.CreatedAt)
	result, _ := s.GetAPI(v.ID)
	return result
}
func (s *PostgresStore) DeleteAPI(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM virtualization_apis WHERE id=$1", id)
	return err == nil && result.RowsAffected() > 0
}
func (s *PostgresStore) ListStubs() []*Stub {
	rows, err := s.pool.Query(context.Background(), "SELECT id,api_id,name,method,endpoint,base_url,environment,category,description,owner_group,request_matcher,response_status,response_body,response_headers,delay_ms,enabled,version FROM virtualization_stubs ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*Stub{}
	for rows.Next() {
		v := &Stub{}
		if rows.Scan(&v.ID, &v.APIID, &v.Name, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.OwnerGroup, &v.RequestMatcher, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.DelayMS, &v.Enabled, &v.Version) == nil {
			out = append(out, v)
		}
	}
	return out
}
func (s *PostgresStore) GetStub(id string) (*Stub, bool) {
	v := &Stub{}
	err := s.pool.QueryRow(context.Background(), "SELECT id,api_id,name,method,endpoint,base_url,environment,category,description,owner_group,request_matcher,response_status,response_body,response_headers,delay_ms,enabled,version FROM virtualization_stubs WHERE id=$1", id).Scan(&v.ID, &v.APIID, &v.Name, &v.Method, &v.Endpoint, &v.BaseURL, &v.Environment, &v.Category, &v.Description, &v.OwnerGroup, &v.RequestMatcher, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.DelayMS, &v.Enabled, &v.Version)
	return v, err == nil
}
func (s *PostgresStore) SaveStub(v *Stub) *Stub {
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
	_, _ = s.pool.Exec(context.Background(), `INSERT INTO virtualization_stubs (id,api_id,name,method,endpoint,base_url,environment,category,description,owner_group,request_matcher,response_status,response_body,response_headers,delay_ms,enabled,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) ON CONFLICT(id) DO UPDATE SET api_id=EXCLUDED.api_id,name=EXCLUDED.name,method=EXCLUDED.method,endpoint=EXCLUDED.endpoint,base_url=EXCLUDED.base_url,environment=EXCLUDED.environment,category=EXCLUDED.category,description=EXCLUDED.description,owner_group=EXCLUDED.owner_group,request_matcher=EXCLUDED.request_matcher,response_status=EXCLUDED.response_status,response_body=EXCLUDED.response_body,response_headers=EXCLUDED.response_headers,delay_ms=EXCLUDED.delay_ms,enabled=EXCLUDED.enabled,version=EXCLUDED.version`, v.ID, v.APIID, v.Name, v.Method, v.Endpoint, v.BaseURL, v.Environment, v.Category, v.Description, v.OwnerGroup, v.RequestMatcher, v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, v.DelayMS, v.Enabled, v.Version)
	
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
func (s *PostgresStore) DeleteStub(id string) bool {
	_, _ = s.pool.Exec(context.Background(), "DELETE FROM virtualization_versions WHERE stub_id=$1", id)
	result, err := s.pool.Exec(context.Background(), "DELETE FROM virtualization_stubs WHERE id=$1", id)
	return err == nil && result.RowsAffected() > 0
}
func (s *PostgresStore) ClearStubs() {
	_, _ = s.pool.Exec(context.Background(), "DELETE FROM virtualization_versions")
	_, _ = s.pool.Exec(context.Background(), "DELETE FROM virtualization_stubs")
}
func (s *PostgresStore) ToggleStub(id string) (*Stub, bool) {
	result, err := s.pool.Exec(context.Background(), "UPDATE virtualization_stubs SET enabled=NOT enabled WHERE id=$1", id)
	if err != nil || result.RowsAffected() == 0 {
		return nil, false
	}
	return s.GetStub(id)
}
func (s *PostgresStore) FindStub(method, url, body string) (*Stub, bool) {
	return s.FindStubForAPI("", method, url, body)
}
func (s *PostgresStore) FindStubForAPI(apiID, method, url, body string) (*Stub, bool) {
	for _, v := range s.ListStubs() {
		if v.Enabled && (apiID == "" || v.APIID == "" || v.APIID == apiID) && methodMatches(v.Method, method) && pathMatches(v.Endpoint, url) && bodyMatches(v.RequestMatcher, body) {
			return v, true
		}
	}
	return nil, false
}
func (s *PostgresStore) ListVersions(stubID string) []*StubVersion {
	rows, err := s.pool.Query(context.Background(), "SELECT id,stub_id,version,version_tag,response_status,response_body,response_headers,active FROM virtualization_versions WHERE stub_id=$1", stubID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*StubVersion{}
	for rows.Next() {
		v := &StubVersion{}
		if rows.Scan(&v.ID, &v.StubID, &v.Version, &v.VersionTag, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.Active) == nil {
			out = append(out, v)
		}
	}
	return out
}
func (s *PostgresStore) SaveVersion(v *StubVersion) *StubVersion {
	if v.ID == "" {
		v.ID = newID()
	}
	_, _ = s.pool.Exec(context.Background(), `INSERT INTO virtualization_versions(id,stub_id,version,version_tag,response_status,response_body,response_headers,active) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO UPDATE SET version=EXCLUDED.version,version_tag=EXCLUDED.version_tag,response_status=EXCLUDED.response_status,response_body=EXCLUDED.response_body,response_headers=EXCLUDED.response_headers,active=EXCLUDED.active`, v.ID, v.StubID, v.Version, v.VersionTag, v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, v.Active)
	if v.Active {
		_, _ = s.pool.Exec(context.Background(), "UPDATE virtualization_stubs SET response_status=$1,response_body=$2,response_headers=$3 WHERE id=$4", v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, v.StubID)
	}
	return v
}
func (s *PostgresStore) DeleteVersion(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM virtualization_versions WHERE id=$1", id)
	return err == nil && result.RowsAffected() > 0
}
func (s *PostgresStore) ActivateVersion(stubID, versionID string) bool {
	tx, err := s.pool.Begin(context.Background())
	if err != nil {
		return false
	}
	defer tx.Rollback(context.Background())
	if _, err = tx.Exec(context.Background(), "UPDATE virtualization_versions SET active=FALSE WHERE stub_id=$1", stubID); err != nil {
		return false
	}
	if _, err = tx.Exec(context.Background(), "UPDATE virtualization_versions SET active=TRUE WHERE id=$1 AND stub_id=$2", versionID, stubID); err != nil {
		return false
	}
	v, ok := s.GetVersion(versionID)
	if !ok {
		return false
	}
	if _, err = tx.Exec(context.Background(), "UPDATE virtualization_stubs SET response_status=$1,response_body=$2,response_headers=$3 WHERE id=$4", v.ResponseStatus, v.ResponseBody, v.ResponseHeaders, stubID); err != nil {
		return false
	}
	return tx.Commit(context.Background()) == nil
}
func (s *PostgresStore) GetVersion(id string) (*StubVersion, bool) {
	v := &StubVersion{}
	err := s.pool.QueryRow(context.Background(), "SELECT id,stub_id,version,version_tag,response_status,response_body,response_headers,active FROM virtualization_versions WHERE id=$1", id).Scan(&v.ID, &v.StubID, &v.Version, &v.VersionTag, &v.ResponseStatus, &v.ResponseBody, &v.ResponseHeaders, &v.Active)
	return v, err == nil
}
func (s *PostgresStore) SaveRequest(v *RequestRecord) *RequestRecord {
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	_, _ = s.pool.Exec(context.Background(), `INSERT INTO virtualization_requests (id,method,url,base_url,endpoint,headers,body,body_s3_key,status,response,response_s3_key,response_headers,is_recorded,category,owner_group,source,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`, v.ID, v.Method, v.URL, v.BaseURL, v.Endpoint, v.Headers, v.Body, v.BodyS3Key, v.Status, v.Response, v.ResponseS3Key, v.ResponseHeaders, v.IsRecorded, v.Category, v.OwnerGroup, v.Source, v.CreatedAt)
	return v
}
func (s *PostgresStore) ListRequests() []*RequestRecord {
	rows, err := s.pool.Query(context.Background(), "SELECT id,method,url,base_url,endpoint,headers,body,body_s3_key,status,response,response_s3_key,response_headers,is_recorded,category,owner_group,source,created_at FROM virtualization_requests ORDER BY created_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := []*RequestRecord{}
	for rows.Next() {
		v := &RequestRecord{}
		if rows.Scan(&v.ID, &v.Method, &v.URL, &v.BaseURL, &v.Endpoint, &v.Headers, &v.Body, &v.BodyS3Key, &v.Status, &v.Response, &v.ResponseS3Key, &v.ResponseHeaders, &v.IsRecorded, &v.Category, &v.OwnerGroup, &v.Source, &v.CreatedAt) == nil {
			out = append(out, v)
		}
	}
	return out
}
func (s *PostgresStore) DeleteRequest(id string) bool {
	result, err := s.pool.Exec(context.Background(), "DELETE FROM virtualization_requests WHERE id=$1", id)
	return err == nil && result.RowsAffected() > 0
}
func (s *PostgresStore) ClearRequests() {
	_, _ = s.pool.Exec(context.Background(), "DELETE FROM virtualization_requests")
}
func (s *PostgresStore) SetSetting(k, v string) {
	_, _ = s.pool.Exec(context.Background(), "INSERT INTO virtualization_settings(key,value) VALUES($1,$2) ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value", k, v)
}
func (s *PostgresStore) Setting(k string) string {
	var v string
	_ = s.pool.QueryRow(context.Background(), "SELECT value FROM virtualization_settings WHERE key=$1", k).Scan(&v)
	return v
}
