package virtualization

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type BlobStore interface {
	Put(key, value string) error
	Get(key string) (string, error)
}

type MemoryBlobStore struct{ values map[string]string }

func NewMemoryBlobStore() *MemoryBlobStore             { return &MemoryBlobStore{values: map[string]string{}} }
func (s *MemoryBlobStore) Put(key, value string) error { s.values[key] = value; return nil }
func (s *MemoryBlobStore) Get(key string) (string, error) {
	value, ok := s.values[key]
	if !ok {
		return "", fmt.Errorf("object %q not found", key)
	}
	return value, nil
}

type ManagementHandler struct{ Store Store }

func (h *ManagementHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/api/registry":
		h.apis(w, r)
	case strings.HasPrefix(path, "/api/registry/"):
		h.api(w, r, strings.TrimPrefix(path, "/api/registry/"))
	case path == "/api/stubs":
		h.stubs(w, r)
	case strings.HasPrefix(path, "/api/stubs/"):
		h.stub(w, r, strings.TrimPrefix(path, "/api/stubs/"))
	case path == "/api/requests" || path == "/api/recorded-requests":
		h.requests(w, r)
	case strings.HasPrefix(path, "/api/requests/"):
		h.request(w, r, strings.TrimPrefix(path, "/api/requests/"))
	case path == "/api/settings":
		h.settings(w, r)
	case strings.HasPrefix(path, "/api/settings/"):
		h.setting(w, r, strings.TrimPrefix(path, "/api/settings/"))
	case path == "/api/analytics":
		h.analytics(w)
	default:
		http.NotFound(w, r)
	}
}
func (h *ManagementHandler) apis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.Store.ListAPIs())
	case http.MethodPost:
		var v API
		if decode(r, &v) != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		writeJSON(w, 200, h.Store.SaveAPI(&v))
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func (h *ManagementHandler) api(w http.ResponseWriter, r *http.Request, id string) {
	switch r.Method {
	case http.MethodGet:
		v, ok := h.Store.GetAPI(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, v)
	case http.MethodPatch, http.MethodPut:
		var v API
		if decode(r, &v) != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		v.ID = id
		writeJSON(w, 200, h.Store.SaveAPI(&v))
	case http.MethodDelete:
		if !h.Store.DeleteAPI(id) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func (h *ManagementHandler) stubs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, h.Store.ListStubs())
	case http.MethodDelete:
		h.Store.ClearStubs()
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		var v Stub
		if decode(r, &v) != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		writeJSON(w, 200, h.Store.SaveStub(&v))
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func (h *ManagementHandler) stub(w http.ResponseWriter, r *http.Request, id string) {
	parts := strings.Split(id, "/")
	stubID := parts[0]
	if len(parts) > 1 && parts[1] == "versions" {
		if len(parts) == 2 && r.Method == http.MethodGet {
			writeJSON(w, 200, h.Store.ListVersions(stubID))
			return
		}
		if len(parts) == 2 && r.Method == http.MethodPost {
			var v StubVersion
			if decode(r, &v) != nil {
				http.Error(w, "invalid body", 400)
				return
			}
			v.StubID = stubID
			writeJSON(w, 200, h.Store.SaveVersion(&v))
			return
		}
		if len(parts) >= 3 {
			versionID := parts[2]
			if len(parts) == 4 && parts[3] == "activate" && r.Method == http.MethodPost {
				if !h.Store.ActivateVersion(stubID, versionID) {
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.Method == http.MethodDelete {
				if !h.Store.DeleteVersion(versionID) {
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				var v StubVersion
				if decode(r, &v) != nil {
					http.Error(w, "invalid body", 400)
					return
				}
				v.ID, v.StubID = versionID, stubID
				writeJSON(w, 200, h.Store.SaveVersion(&v))
				return
			}
		}
		http.Error(w, "method not allowed", 405)
		return
	}
	switch r.Method {
	case http.MethodGet:
		v, ok := h.Store.GetStub(stubID)
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, v)
	case http.MethodPut, http.MethodPatch:
		var v Stub
		if decode(r, &v) != nil {
			http.Error(w, "invalid body", 400)
			return
		}
		v.ID = stubID
		writeJSON(w, 200, h.Store.SaveStub(&v))
	case http.MethodDelete:
		if !h.Store.DeleteStub(stubID) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodPost:
		if len(parts) == 2 && parts[1] == "toggle" {
			v, ok := h.Store.ToggleStub(stubID)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, http.StatusOK, v)
			return
		}
	default:
		http.Error(w, "method not allowed", 405)
	}
}
func (h *ManagementHandler) analytics(w http.ResponseWriter) {
	writeJSON(w, 200, map[string]int{"apis": len(h.Store.ListAPIs()), "stubs": len(h.Store.ListStubs()), "requests": len(h.Store.ListRequests())})
}

func (h *ManagementHandler) requests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, h.Store.ListRequests())
	case http.MethodDelete:
		h.Store.ClearRequests()
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ManagementHandler) request(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.Store.DeleteRequest(id) {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManagementHandler) settings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	settings := map[string]string{"recordingMode": "true", "playbackMode": "true", "targetUrl": "", "useToggle": "true", "virtualizationEnabled": "true"}
	for _, key := range []string{"recordingMode", "playbackMode", "targetUrl", "useToggle", "virtualizationEnabled"} {
		if value := h.Store.Setting(key); value != "" {
			settings[key] = value
		}
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *ManagementHandler) setting(w http.ResponseWriter, r *http.Request, key string) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if decode(r, &body) != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.Store.SetSetting(key, body.Value)
	w.WriteHeader(http.StatusNoContent)
}

func identifyAPI(apis []*API, method, path string) *API {
	for _, api := range apis {
		if api.Enabled && methodMatches(api.Method, method) && pathMatches(api.Endpoint, path) {
			return api
		}
	}
	return nil
}

type RuntimeHandler struct {
	Store Store
	Blobs BlobStore
}

func (h *RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	body := string(bodyBytes)
	if r.URL.Path == "/api/proxy-request" || r.URL.Path == "/api/proxy-external" {
		h.proxy(w, r, body)
		return
	}
	if h.Store.Setting("virtualizationEnabled") == "false" {
		http.Error(w, `{"error":"virtualization is disabled"}`, http.StatusServiceUnavailable)
		return
	}
	apiID := ""
	apis := h.Store.ListAPIs()
	if len(apis) > 0 {
		api := identifyAPI(apis, r.Method, r.URL.Path)
		if api == nil {
			http.Error(w, `{"error":"API is not registered or disabled"}`, http.StatusNotFound)
			return
		}
		apiID = api.ID
	}
	stub, ok := h.Store.FindStubForAPI(apiID, r.Method, r.URL.Path, body)
	if !ok {
		http.Error(w, `{"error":"no matching stub"}`, http.StatusNotFound)
		return
	}
	if stub.DelayMS > 0 {
		time.Sleep(time.Duration(stub.DelayMS) * time.Millisecond)
	}
	responseBody := stub.ResponseBody
	if strings.HasPrefix(responseBody, "s3://") {
		value, err := h.Blobs.Get(strings.TrimPrefix(responseBody, "s3://"))
		if err != nil {
			http.Error(w, `{"error":"stub response unavailable"}`, 500)
			return
		}
		responseBody = value
	}
	applyHeaders(w, stub.ResponseHeaders)
	if stub.ResponseStatus < 100 || stub.ResponseStatus > 599 {
		stub.ResponseStatus = 200
	}
	w.Header().Set("X-Response-Source", "stub")
	w.WriteHeader(stub.ResponseStatus)
	_, _ = w.Write([]byte(responseBody))
}

func (h *RuntimeHandler) proxy(w http.ResponseWriter, r *http.Request, body string) {
	endpoint := r.URL.Query().Get("endpoint")
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		targetURL = strings.TrimRight(r.Header.Get("X-Target-Host"), "/") + "/" + strings.TrimLeft(endpoint, "/")
	}
	if targetURL == "" || targetURL == "/" {
		http.Error(w, `{"error":"URL or target host is required"}`, http.StatusBadRequest)
		return
	}
	mockMode := strings.EqualFold(h.Store.Setting("useToggle"), "off") || strings.EqualFold(h.Store.Setting("playbackMode"), "true") && strings.EqualFold(h.Store.Setting("useToggle"), "mock")
	parsed, err := http.NewRequest(r.Method, targetURL, strings.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"invalid target URL"}`, http.StatusBadRequest)
		return
	}
	for key, values := range r.Header {
		if len(values) > 0 && !strings.EqualFold(key, "Host") {
			parsed.Header.Set(key, values[0])
		}
	}
	if mockMode {
		if h.Store.Setting("virtualizationEnabled") == "false" {
			http.Error(w, `{"error":"virtualization is disabled"}`, http.StatusServiceUnavailable)
			return
		}
		stub, ok := h.Store.FindStubForAPI("", r.Method, parsed.URL.Path, body)
		if !ok {
			w.Header().Set("X-No-Stub-Found", "true")
			http.Error(w, `{"error":"no matching stub"}`, http.StatusNotFound)
			return
		}
		h.writeStubResponse(w, stub)
		return
	}
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(parsed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	for key, values := range response.Header {
		if len(values) > 0 {
			w.Header().Set(key, values[0])
		}
	}
	w.Header().Set("X-Response-Source", "live-api")
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(responseBody)
}

func (h *RuntimeHandler) writeStubResponse(w http.ResponseWriter, stub *Stub) {
	if stub.DelayMS > 0 {
		time.Sleep(time.Duration(stub.DelayMS) * time.Millisecond)
	}
	responseBody := stub.ResponseBody
	if strings.HasPrefix(responseBody, "s3://") {
		value, err := h.Blobs.Get(strings.TrimPrefix(responseBody, "s3://"))
		if err != nil {
			http.Error(w, `{"error":"stub response unavailable"}`, http.StatusInternalServerError)
			return
		}
		responseBody = value
	}
	applyHeaders(w, stub.ResponseHeaders)
	if stub.ResponseStatus < 100 || stub.ResponseStatus > 599 {
		stub.ResponseStatus = http.StatusOK
	}
	w.Header().Set("X-Response-Source", "stub")
	w.WriteHeader(stub.ResponseStatus)
	_, _ = w.Write([]byte(responseBody))
}

type CaptureHandler struct {
	Store  Store
	Blobs  BlobStore
	Client *http.Client
}

type HealthCheckHandler struct{ Client *http.Client }

func (h *HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Params  map[string]string `json:"params"`
		Body    string            `json:"body"`
	}
	if decode(r, &input) != nil || input.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}
	if input.Method == "" {
		input.Method = http.MethodGet
	}
	targetURL := input.URL
	if len(input.Params) > 0 {
		query := ""
		for key, value := range input.Params {
			if query != "" {
				query += "&"
			}
			query += url.QueryEscape(key) + "=" + url.QueryEscape(value)
		}
		if strings.Contains(targetURL, "?") {
			targetURL += "&" + query
		} else {
			targetURL += "?" + query
		}
	}
	request, err := http.NewRequest(input.Method, targetURL, strings.NewReader(input.Body))
	if err != nil {
		http.Error(w, "invalid target URL", http.StatusBadRequest)
		return
	}
	for key, value := range input.Headers {
		request.Header.Set(key, value)
	}
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	started := time.Now()
	response, err := client.Do(request)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"error": err.Error()})
		return
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	headers := map[string]string{}
	for key, values := range response.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"statusCode": response.StatusCode, "responseTime": time.Since(started).Milliseconds(), "body": string(responseBody), "headers": headers})
}

func (h *CaptureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var input CaptureRequest
	if decode(r, &input) != nil || input.URL == "" {
		http.Error(w, "url is required", 400)
		return
	}
	if input.Method == "" {
		input.Method = http.MethodGet
	}
	req, err := http.NewRequest(input.Method, input.URL, strings.NewReader(input.Body))
	if err != nil {
		http.Error(w, "invalid target url", 400)
		return
	}
	for k, v := range input.Headers {
		req.Header.Set(k, v)
	}
	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	response, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 502)
		return
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	responseHeaders := map[string]string{}
	for k, values := range response.Header {
		if len(values) > 0 {
			responseHeaders[k] = values[0]
		}
	}
	record := &RequestRecord{Method: input.Method, URL: input.URL, Endpoint: req.URL.Path, Body: input.Body, Status: response.StatusCode, Response: string(responseBody), ResponseHeaders: encode(responseHeaders), Source: "capture"}
	if input.Body != "" {
		record.BodyS3Key = "requests/" + newID() + "/request"
		_ = h.Blobs.Put(record.BodyS3Key, input.Body)
	}
	record.ResponseS3Key = "requests/" + newID() + "/response"
	_ = h.Blobs.Put(record.ResponseS3Key, record.Response)
	saved := h.Store.SaveRequest(record)
	apiID := ""
	if api := identifyAPI(h.Store.ListAPIs(), input.Method, req.URL.Path); api != nil {
		apiID = api.ID
	}
	autoStub := h.Store.SaveStub(&Stub{
		APIID:           apiID,
		Name:            "Captured " + input.Method + " " + req.URL.Path,
		Method:          input.Method,
		Endpoint:        req.URL.Path,
		RequestMatcher:  input.Body,
		ResponseStatus:  response.StatusCode,
		ResponseBody:    "s3://" + record.ResponseS3Key,
		ResponseHeaders: encode(responseHeaders),
		Enabled:         true,
	})
	writeJSON(w, 200, map[string]interface{}{"request": saved, "stub": autoStub, "response": CaptureResponse{Status: response.StatusCode, Headers: responseHeaders, Body: string(responseBody)}})
}

func decode(r *http.Request, target interface{}) error { return json.NewDecoder(r.Body).Decode(target) }
func encode(v interface{}) string                      { data, _ := json.Marshal(v); return string(data) }
func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func applyHeaders(w http.ResponseWriter, raw string) {
	var headers map[string]string
	if json.Unmarshal([]byte(raw), &headers) != nil {
		return
	}
	for key, value := range headers {
		lower := strings.ToLower(key)
		if lower != "content-length" && lower != "connection" && lower != "transfer-encoding" {
			w.Header().Set(key, value)
		}
	}
}
