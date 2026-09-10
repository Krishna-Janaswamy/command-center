package virtualization

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeReturnsMatchingStub(t *testing.T) {
	store := NewMemoryStore()
	store.SaveStub(&Stub{Method: "POST", Endpoint: "/claim", Enabled: true, ResponseStatus: 201, ResponseBody: `{"approved":true}`})
	req := httptest.NewRequest(http.MethodPost, "/claim", strings.NewReader(`{"id":"123"}`))
	recorder := httptest.NewRecorder()
	(&RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}).ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated || recorder.Body.String() != `{"approved":true}` {
		t.Fatalf("unexpected runtime response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestRuntimeDoesNotServeDisabledStub(t *testing.T) {
	store := NewMemoryStore()
	store.SaveStub(&Stub{Method: "POST", Endpoint: "/claim", Enabled: false, ResponseBody: `{"approved":true}`})
	recorder := httptest.NewRecorder()
	(&RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/claim", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", recorder.Code)
	}
}

func TestCaptureStoresTargetResponseAndMetadata(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Target", "yes")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer target.Close()
	store, blobs := NewMemoryStore(), NewMemoryBlobStore()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/capture", strings.NewReader(`{"method":"GET","url":"`+target.URL+`/claim","body":"payload"}`))
	(&CaptureHandler{Store: store, Blobs: blobs}).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || len(store.ListRequests()) != 1 {
		t.Fatalf("capture failed: %d records=%d body=%s", recorder.Code, len(store.ListRequests()), recorder.Body.String())
	}
	record := store.ListRequests()[0]
	if record.Status != http.StatusAccepted || record.ResponseS3Key == "" {
		t.Fatalf("metadata not stored: %+v", record)
	}
	if _, err := blobs.Get(record.ResponseS3Key); err != nil {
		t.Fatalf("response was not stored: %v", err)
	}
	stubs := store.ListStubs()
	if len(stubs) != 1 || stubs[0].ResponseBody != "s3://"+record.ResponseS3Key {
		t.Fatalf("captured response did not create an S3-backed stub: %+v", stubs)
	}
}

func TestSaveStubCreatesInitialVersionAndPreservesMetadata(t *testing.T) {
	store := NewMemoryStore()
	stub := store.SaveStub(&Stub{
		Name:            "Get Claims",
		Method:          "GET",
		Endpoint:        "/api/claims",
		Category:        "Claims",
		Environment:     "Dev",
		Description:     "Retrieve claims list",
		ResponseStatus:  200,
		ResponseBody:    `{"claims":[]}`,
		ResponseHeaders: `{"Content-Type":"application/json"}`,
	})
	if stub.ID == "" || stub.Category != "Claims" || stub.Environment != "Dev" || stub.Description != "Retrieve claims list" {
		t.Fatalf("stub fields not saved correctly: %+v", stub)
	}
	versions := store.ListVersions(stub.ID)
	if len(versions) != 1 {
		t.Fatalf("expected 1 initial version, got %d", len(versions))
	}
	if versions[0].Version != "v1" || !versions[0].Active || versions[0].ResponseStatus != 200 {
		t.Fatalf("initial version mismatch: %+v", versions[0])
	}
}

func TestServeMockEndpoint(t *testing.T) {
	store := NewMemoryStore()
	stub := store.SaveStub(&Stub{
		Name:           "Mock Test",
		Method:         "POST",
		Endpoint:       "/api/test",
		ResponseStatus: 202,
		ResponseBody:   `{"status":"accepted"}`,
		Enabled:        true,
	})

	handler := &ManagementHandler{Store: store}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/mock/"+stub.ID, nil)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "accepted") {
		t.Fatalf("unexpected mock body: %s", recorder.Body.String())
	}
}

func TestProxyLiveApiAutoCreatesStub(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"live":true}`))
	}))
	defer target.Close()

	store := NewMemoryStore()
	handler := &RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}

	req := httptest.NewRequest(http.MethodGet, "/api/proxy-request?url="+target.URL+"/user&endpoint=/user", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	stubs := store.ListStubs()
	if len(stubs) != 1 {
		t.Fatalf("expected 1 auto-created stub, got %d", len(stubs))
	}
	if stubs[0].Endpoint != "/user" || stubs[0].ResponseBody != `{"live":true}` {
		t.Fatalf("auto-created stub content mismatch: %+v", stubs[0])
	}
}

func TestApiRegistryFiltering(t *testing.T) {
	store := NewMemoryStore()
	store.SaveAPI(&API{Name: "Claims API", Environment: "Dev", Category: "Claims"})
	store.SaveAPI(&API{Name: "SBI API", Environment: "Prod", Category: "SBI"})

	handler := &ManagementHandler{Store: store}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/registry?env=Dev&category=Claims", nil)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Claims API") || strings.Contains(recorder.Body.String(), "SBI API") {
		t.Fatalf("unexpected API filtering result: %s", recorder.Body.String())
	}
}

func TestSQLiteStorePersistence(t *testing.T) {
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create sqlite store: %v", err)
	}

	api := store.SaveAPI(&API{Name: "Test API", Method: "GET", Endpoint: "/api/test", Environment: "Dev", Category: "Other"})
	if api.ID == "" || len(store.ListAPIs()) != 1 {
		t.Fatalf("failed to save API in sqlite: %+v", api)
	}

	stub := store.SaveStub(&Stub{Name: "Test Stub", Method: "GET", Endpoint: "/api/test", ResponseStatus: 200, ResponseBody: `{"ok":true}`})
	if stub.ID == "" || len(store.ListStubs()) != 1 {
		t.Fatalf("failed to save Stub in sqlite: %+v", stub)
	}

	versions := store.ListVersions(stub.ID)
	if len(versions) != 1 || !versions[0].Active {
		t.Fatalf("failed to get initial version in sqlite: %+v", versions)
	}
}

func TestProxyToggleOffMissingStubHitsLiveAndCreatesStub(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"autoCreatedFromLive":true}`))
	}))
	defer target.Close()

	store := NewMemoryStore()
	store.SetSetting("useToggle", "off")
	handler := &RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}

	// First request: toggle OFF, no existing stub. Should hit live API and auto-create stub.
	req := httptest.NewRequest(http.MethodGet, "/api/proxy-request?url="+target.URL+"/orders&endpoint=/orders", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != `{"autoCreatedFromLive":true}` {
		t.Fatalf("expected live API response, got %d %s", rec.Code, rec.Body.String())
	}

	stubs := store.ListStubs()
	if len(stubs) != 1 {
		t.Fatalf("expected 1 auto-created stub, got %d", len(stubs))
	}
	if stubs[0].Endpoint != "/orders" || stubs[0].ResponseBody != `{"autoCreatedFromLive":true}` {
		t.Fatalf("stub was not auto-created properly: %+v", stubs[0])
	}

	// Second request: toggle OFF, now stub exists! Should return stub response.
	req2 := httptest.NewRequest(http.MethodGet, "/api/proxy-request?url="+target.URL+"/orders&endpoint=/orders", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK || rec2.Body.String() != `{"autoCreatedFromLive":true}` {
		t.Fatalf("expected stub response on second call, got %d %s", rec2.Code, rec2.Body.String())
	}
	if rec2.Header().Get("X-Response-Source") != "stub" {
		t.Fatalf("expected X-Response-Source header to be stub, got %s", rec2.Header().Get("X-Response-Source"))
	}
}

func TestProxyToggleOffAutoStubEvenWhenRecordingDisabled(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"forced":true}`))
	}))
	defer target.Close()

	store := NewMemoryStore()
	store.SetSetting("useToggle", "off")
	store.SetSetting("recordingMode", "false")
	handler := &RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}

	req := httptest.NewRequest(http.MethodGet, "/api/proxy-request?url="+target.URL+"/items&endpoint=/items", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected live response, got %d", rec.Code)
	}
	if len(store.ListStubs()) != 1 {
		t.Fatalf("expected auto-stub even with recordingMode=false, got %d stubs", len(store.ListStubs()))
	}
}

func TestCatchAllToggleOffMissProxiesRegisteredAPIAndCreatesStub(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/claims" {
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"claim":"ok"}`))
	}))
	defer target.Close()

	store := NewMemoryStore()
	store.SetSetting("useToggle", "off")
	store.SaveAPI(&API{
		Name:     "Claims",
		Method:   "GET",
		Endpoint: "/v1/claims",
		BaseURL:  target.URL,
		Enabled:  true,
	})
	handler := &RuntimeHandler{Store: store, Blobs: NewMemoryBlobStore()}

	req := httptest.NewRequest(http.MethodGet, "/v1/claims", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || rec.Body.String() != `{"claim":"ok"}` {
		t.Fatalf("expected live passthrough, got %d %s", rec.Code, rec.Body.String())
	}
	stubs := store.ListStubs()
	if len(stubs) != 1 || stubs[0].Endpoint != "/v1/claims" {
		t.Fatalf("expected auto-created stub for catch-all, got %+v", stubs)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/claims", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Header().Get("X-Response-Source") != "stub" {
		t.Fatalf("expected stub on second catch-all call, got %s", rec2.Header().Get("X-Response-Source"))
	}
}

// mockTransport lets tests control what the proxy's HTTP client returns.
type mockTransport struct {
	handler http.Handler
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	return rec.Result(), nil
}

func mockClient(h http.Handler) *http.Client {
	return &http.Client{Transport: &mockTransport{handler: h}}
}

// TestStubExistsHitLiveAPI_StubPreserved is the core scenario the user reported:
//   1. A manual stub exists for /api/orders (returns 200 stub body).
//   2. User hits the live API (allowRealApi=true) from the API Tester.
//   3. The live response MUST be returned (not an error).
//   4. The existing stub MUST NOT be overwritten — it should still return the original stub body.
//   5. A new inactive version capturing the live response MUST be recorded.
func TestStubExistsHitLiveAPI_StubPreserved(t *testing.T) {
	store := NewMemoryStore()
	store.SetSetting("playbackMode", "true")
	store.SetSetting("useToggle", "mock")

	// Pre-existing manual stub.
	stub := store.SaveStub(&Stub{
		Method:         "GET",
		Endpoint:       "/api/orders",
		ResponseStatus: 200,
		ResponseBody:   `{"source":"stub","orders":[]}`,
		Enabled:        true,
	})

	// Simulate a live upstream that returns a different body.
	liveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"source":"live","orders":[{"id":1}]}`))
	})

	handler := &RuntimeHandler{
		Store:  store,
		Blobs:  NewMemoryBlobStore(),
		Client: mockClient(liveHandler),
	}

	// Step 1: Normal request should serve the stub.
	req1 := httptest.NewRequest(http.MethodGet, "/api/proxy-request?endpoint=/api/orders&url=http://upstream/api/orders", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != 200 {
		t.Fatalf("step1: expected 200 from stub, got %d", rec1.Code)
	}
	if rec1.Header().Get("X-Response-Source") != "stub" {
		t.Fatalf("step1: expected stub source, got %q body=%s", rec1.Header().Get("X-Response-Source"), rec1.Body.String())
	}
	if rec1.Body.String() != `{"source":"stub","orders":[]}` {
		t.Fatalf("step1: expected stub body, got %s", rec1.Body.String())
	}

	// Step 2: Hit live API (allowRealApi=true, as the frontend sends after user confirms).
	req2 := httptest.NewRequest(http.MethodGet, "/api/proxy-request?endpoint=/api/orders&url=http://upstream/api/orders&allowRealApi=true", nil)
	req2.Header.Set("X-Target-Host", "http://upstream")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != 200 {
		t.Fatalf("step2: expected 200 from live API, got %d body=%s", rec2.Code, rec2.Body.String())
	}
	if rec2.Header().Get("X-Response-Source") != "live-api" {
		t.Fatalf("step2: expected live-api source, got %q", rec2.Header().Get("X-Response-Source"))
	}
	if rec2.Body.String() != `{"source":"live","orders":[{"id":1}]}` {
		t.Fatalf("step2: expected live body, got %s", rec2.Body.String())
	}

	// Step 3: The original stub MUST still be intact (not overwritten by the live response).
	refreshed, ok := store.GetStub(stub.ID)
	if !ok {
		t.Fatal("step3: stub was deleted after hitting live API")
	}
	if refreshed.ResponseBody != `{"source":"stub","orders":[]}` {
		t.Fatalf("step3: stub body was overwritten! got %s", refreshed.ResponseBody)
	}

	// Step 4: A new inactive version capturing the live response must have been recorded.
	versions := store.ListVersions(stub.ID)
	// Should have at least 2 versions: initial active + new inactive auto-recorded.
	if len(versions) < 2 {
		t.Fatalf("step4: expected at least 2 versions (initial + auto-recorded), got %d", len(versions))
	}
	foundLiveVersion := false
	foundActiveVersion := false
	for _, v := range versions {
		if v.ResponseBody == `{"source":"live","orders":[{"id":1}]}` && !v.Active {
			foundLiveVersion = true
		}
		if v.Active && v.ResponseBody == `{"source":"stub","orders":[]}` {
			foundActiveVersion = true
		}
	}
	if !foundLiveVersion {
		t.Fatal("step4: no inactive version with live response body was recorded")
	}
	if !foundActiveVersion {
		t.Fatal("step4: active version no longer has the original stub body")
	}

	// Step 5: Subsequent normal requests must still return the original stub.
	req3 := httptest.NewRequest(http.MethodGet, "/api/proxy-request?endpoint=/api/orders&url=http://upstream/api/orders", nil)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Header().Get("X-Response-Source") != "stub" {
		t.Fatalf("step5: expected stub source after live hit, got %q", rec3.Header().Get("X-Response-Source"))
	}
	if rec3.Body.String() != `{"source":"stub","orders":[]}` {
		t.Fatalf("step5: stub response was corrupted after live hit, got %s", rec3.Body.String())
	}
}

// TestStubExistsHitLiveAPI_NoCORSHeaderLeak verifies that CORS headers from the live
// upstream do NOT leak into the proxied response (which would conflict with our backend CORS).
func TestStubExistsHitLiveAPI_NoCORSHeaderLeak(t *testing.T) {
	store := NewMemoryStore()

	liveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://evil.com")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	handler := &RuntimeHandler{
		Store:  store,
		Blobs:  NewMemoryBlobStore(),
		Client: mockClient(liveHandler),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/proxy-request?endpoint=/data&url=http://upstream/data&allowRealApi=true", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	if acao := rec.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Fatalf("upstream CORS header must not leak to response, got: %s", acao)
	}
}

// TestStubHitIsRecordedInRequests verifies that when a stub serves a request,
// it is saved in the store with source="stub" so it appears in Recorded Requests.
func TestStubHitIsRecordedInRequests(t *testing.T) {
	store := NewMemoryStore()
	store.SetSetting("playbackMode", "true")
	store.SetSetting("useToggle", "mock")

	store.SaveStub(&Stub{
		Method:         "GET",
		Endpoint:       "/api/claims",
		ResponseStatus: 200,
		ResponseBody:   `{"claims":[]}`,
		Category:       "Claims",
		Enabled:        true,
	})

	handler := &RuntimeHandler{
		Store: store,
		Blobs: NewMemoryBlobStore(),
	}

	// Hit via /api/proxy-request (the API Tester path)
	req := httptest.NewRequest(http.MethodGet, "/api/proxy-request?endpoint=/api/claims&url=http://upstream/api/claims", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Response-Source") != "stub" {
		t.Fatalf("expected stub source, got %q", rec.Header().Get("X-Response-Source"))
	}

	// The stub hit must be recorded.
	requests := store.ListRequests()
	if len(requests) != 1 {
		t.Fatalf("expected 1 recorded request, got %d", len(requests))
	}
	r := requests[0]
	if r.Source != "stub" {
		t.Fatalf("expected source=stub, got %q", r.Source)
	}
	if r.Status != 200 {
		t.Fatalf("expected status=200, got %d", r.Status)
	}
	if r.Response != `{"claims":[]}` {
		t.Fatalf("expected claims response body, got %q", r.Response)
	}
	if r.Category != "Claims" {
		t.Fatalf("expected category=Claims, got %q", r.Category)
	}
}
