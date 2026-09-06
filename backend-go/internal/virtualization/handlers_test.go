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
