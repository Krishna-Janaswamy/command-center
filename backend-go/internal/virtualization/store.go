package virtualization

import (
	"strings"
	"sync"
	"time"
)

type Store interface {
	ListAPIs() []*API
	GetAPI(string) (*API, bool)
	SaveAPI(*API) *API
	DeleteAPI(string) bool
	ListStubs() []*Stub
	GetStub(string) (*Stub, bool)
	SaveStub(*Stub) *Stub
	DeleteStub(string) bool
	ClearStubs()
	ToggleStub(string) (*Stub, bool)
	FindStub(method, url, body string) (*Stub, bool)
	FindStubForAPI(apiID, method, url, body string) (*Stub, bool)
	ListVersions(string) []*StubVersion
	SaveVersion(*StubVersion) *StubVersion
	DeleteVersion(string) bool
	ActivateVersion(string, string) bool
	SaveRequest(*RequestRecord) *RequestRecord
	ListRequests() []*RequestRecord
	DeleteRequest(string) bool
	ClearRequests()
	SetSetting(string, string)
	Setting(string) string
}

type MemoryStore struct {
	versions map[string]*StubVersion
	mu       sync.RWMutex
	apis     map[string]*API
	stubs    map[string]*Stub
	requests []*RequestRecord
	settings map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{apis: map[string]*API{}, stubs: map[string]*Stub{}, settings: map[string]string{}, versions: map[string]*StubVersion{}}
}
func (s *MemoryStore) ListAPIs() []*API {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*API, 0, len(s.apis))
	for _, v := range s.apis {
		out = append(out, cloneAPI(v))
	}
	return out
}
func (s *MemoryStore) GetAPI(id string) (*API, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.apis[id]
	return cloneAPI(v), ok
}
func (s *MemoryStore) SaveAPI(v *API) *API {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	if !v.Enabled {
		v.Enabled = true
	}
	s.apis[v.ID] = cloneAPI(v)
	return cloneAPI(v)
}
func (s *MemoryStore) DeleteAPI(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.apis[id]; !ok {
		return false
	}
	delete(s.apis, id)
	return true
}
func (s *MemoryStore) ListStubs() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Stub, 0, len(s.stubs))
	for _, v := range s.stubs {
		out = append(out, cloneStub(v))
	}
	return out
}
func (s *MemoryStore) GetStub(id string) (*Stub, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.stubs[id]
	return cloneStub(v), ok
}
func (s *MemoryStore) SaveStub(v *Stub) *Stub {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.ID == "" {
		v.ID = newID()
	}
	if v.Method == "" {
		v.Method = "ANY"
	}
	if v.Version == "" {
		v.Version = "v1"
	}
	s.stubs[v.ID] = cloneStub(v)
	return cloneStub(v)
}
func (s *MemoryStore) DeleteStub(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.stubs[id]; !ok {
		return false
	}
	delete(s.stubs, id)
	return true
}
func (s *MemoryStore) ClearStubs() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stubs = map[string]*Stub{}
	s.versions = map[string]*StubVersion{}
}
func (s *MemoryStore) ToggleStub(id string) (*Stub, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.stubs[id]
	if !ok {
		return nil, false
	}
	v.Enabled = !v.Enabled
	return cloneStub(v), true
}
func (s *MemoryStore) FindStub(method, url, body string) (*Stub, bool) {
	return s.FindStubForAPI("", method, url, body)
}
func (s *MemoryStore) FindStubForAPI(apiID, method, url, body string) (*Stub, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.stubs {
		if !v.Enabled || (apiID != "" && v.APIID != "" && v.APIID != apiID) || !methodMatches(v.Method, method) || !pathMatches(v.Endpoint, url) || !bodyMatches(v.RequestMatcher, body) {
			continue
		}
		return cloneStub(v), true
	}
	return nil, false
}
func (s *MemoryStore) ListVersions(stubID string) []*StubVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*StubVersion{}
	for _, v := range s.versions {
		if v.StubID == stubID {
			c := *v
			out = append(out, &c)
		}
	}
	return out
}
func (s *MemoryStore) SaveVersion(v *StubVersion) *StubVersion {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.ID == "" {
		v.ID = newID()
	}
	c := *v
	s.versions[v.ID] = &c
	return &c
}
func (s *MemoryStore) DeleteVersion(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.versions[id]; !ok {
		return false
	}
	delete(s.versions, id)
	return true
}
func (s *MemoryStore) ActivateVersion(stubID, versionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	target, ok := s.versions[versionID]
	if !ok || target.StubID != stubID {
		return false
	}
	for _, v := range s.versions {
		if v.StubID == stubID {
			v.Active = false
		}
	}
	target.Active = true
	if stub, ok := s.stubs[stubID]; ok {
		stub.ResponseStatus = target.ResponseStatus
		stub.ResponseBody = target.ResponseBody
		stub.ResponseHeaders = target.ResponseHeaders
	}
	return true
}
func (s *MemoryStore) SaveRequest(v *RequestRecord) *RequestRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v.ID == "" {
		v.ID = newID()
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now().UTC()
	}
	s.requests = append(s.requests, cloneRequest(v))
	return cloneRequest(v)
}
func (s *MemoryStore) ListRequests() []*RequestRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*RequestRecord, len(s.requests))
	for i, v := range s.requests {
		out[i] = cloneRequest(v)
	}
	return out
}
func (s *MemoryStore) DeleteRequest(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.requests {
		if v.ID == id {
			s.requests = append(s.requests[:i], s.requests[i+1:]...)
			return true
		}
	}
	return false
}
func (s *MemoryStore) ClearRequests() { s.mu.Lock(); defer s.mu.Unlock(); s.requests = nil }
func (s *MemoryStore) SetSetting(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings[key] = value
}
func (s *MemoryStore) Setting(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings[key]
}
func methodMatches(expected, actual string) bool {
	return expected == "" || strings.EqualFold(expected, "ANY") || strings.EqualFold(expected, actual)
}
func pathMatches(pattern, actual string) bool {
	return pattern == "" || pattern == actual || strings.Contains(actual, pattern) || strings.HasSuffix(pattern, "*") && strings.HasPrefix(actual, strings.TrimSuffix(pattern, "*"))
}
func bodyMatches(matcher, body string) bool {
	return matcher == "" || matcher == body || strings.Contains(body, matcher)
}
func newID() string { return time.Now().UTC().Format("20060102150405.000000000") }
func cloneAPI(v *API) *API {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
func cloneStub(v *Stub) *Stub {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
func cloneRequest(v *RequestRecord) *RequestRecord {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
