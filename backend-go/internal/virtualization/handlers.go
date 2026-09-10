package virtualization

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
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
	case path == "/api/mock" || strings.HasPrefix(path, "/api/mock/"):
		h.mock(w, r)
	case path == "/api/analytics":
		h.analytics(w)
	default:
		http.NotFound(w, r)
	}
}
func (h *ManagementHandler) apis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		envFilter := r.URL.Query().Get("env")
		categoryFilter := r.URL.Query().Get("category")
		apis := h.Store.ListAPIs()
		if envFilter != "" || categoryFilter != "" {
			filtered := []*API{}
			for _, api := range apis {
				if envFilter != "" && !strings.EqualFold(api.Environment, envFilter) {
					continue
				}
				if categoryFilter != "" && !strings.EqualFold(api.Category, categoryFilter) {
					continue
				}
				filtered = append(filtered, api)
			}
			apis = filtered
		}
		writeJSON(w, http.StatusOK, apis)
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
func (h *ManagementHandler) mock(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/mock")
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"stub id is required"}`, http.StatusBadRequest)
		return
	}
	stubID := parts[0]
	stub, ok := h.Store.GetStub(stubID)
	if !ok || !stub.Enabled {
		http.Error(w, `{"error":"Mock endpoint not found or disabled."}`, http.StatusNotFound)
		return
	}
	if stub.DelayMS > 0 {
		time.Sleep(time.Duration(stub.DelayMS) * time.Millisecond)
	}
	responseBody := stub.ResponseBody
	applyHeaders(w, stub.ResponseHeaders)
	if stub.ResponseStatus < 100 || stub.ResponseStatus > 599 {
		stub.ResponseStatus = http.StatusOK
	}
	w.Header().Set("X-Response-Source", "stub")
	w.WriteHeader(stub.ResponseStatus)
	_, _ = w.Write([]byte(responseBody))
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
	Store  Store
	Blobs  BlobStore
	Client *http.Client // injectable for testing; nil uses default
}

func isMockMode(store Store) bool {
	useToggle := store.Setting("useToggle")
	return strings.EqualFold(useToggle, "off") ||
		(strings.EqualFold(store.Setting("playbackMode"), "true") && strings.EqualFold(useToggle, "mock"))
}

func (h *RuntimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, _ := io.ReadAll(r.Body)
	body := string(bodyBytes)
	if strings.HasPrefix(r.URL.Path, "/api/mock/") || r.URL.Path == "/api/mock" {
		h.serveMock(w, r)
		return
	}
	if r.URL.Path == "/api/proxy-request" || r.URL.Path == "/api/proxy-external" {
		h.proxy(w, r, body)
		return
	}
	if h.Store.Setting("virtualizationEnabled") == "false" {
		http.Error(w, `{"error":"virtualization is disabled"}`, http.StatusServiceUnavailable)
		return
	}
	apiID := ""
	var matchedAPI *API
	apis := h.Store.ListAPIs()
	if len(apis) > 0 {
		matchedAPI = identifyAPI(apis, r.Method, r.URL.Path)
		if matchedAPI == nil {
			http.Error(w, `{"error":"API is not registered or disabled"}`, http.StatusNotFound)
			return
		}
		apiID = matchedAPI.ID
	}
	stub, ok := h.Store.FindStubForAPI(apiID, r.Method, r.URL.Path, body)
	mockMode := isMockMode(h.Store)
	if ok && mockMode {
		h.writeStubResponse(w, r, body, r.URL.Path, stub)
		return
	}
	// Toggle OFF + no stub (or LIVE mode): hit registered upstream and auto-create stub.
	if matchedAPI != nil && strings.TrimSpace(matchedAPI.BaseURL) != "" {
		target := strings.TrimRight(matchedAPI.BaseURL, "/") + "/" + strings.TrimLeft(r.URL.Path, "/")
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		req := r.Clone(r.Context())
		req.URL, _ = url.Parse("/api/proxy-request")
		q := req.URL.Query()
		q.Set("url", target)
		q.Set("endpoint", r.URL.Path)
		req.URL.RawQuery = q.Encode()
		if req.Header.Get("X-Target-Host") == "" {
			req.Header.Set("X-Target-Host", strings.TrimRight(matchedAPI.BaseURL, "/"))
		}
		h.proxy(w, req, body)
		return
	}
	if !ok {
		http.Error(w, `{"error":"no matching stub"}`, http.StatusNotFound)
		return
	}
	h.writeStubResponse(w, r, body, r.URL.Path, stub)
}

func (h *RuntimeHandler) serveMock(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/mock")
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"stub id is required"}`, http.StatusBadRequest)
		return
	}
	stubID := parts[0]
	stub, ok := h.Store.GetStub(stubID)
	if !ok || !stub.Enabled {
		http.Error(w, `{"error":"Mock endpoint not found or disabled."}`, http.StatusNotFound)
		return
	}
	h.writeStubResponse(w, r, "", "", stub)
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
	forceLive := strings.EqualFold(r.URL.Query().Get("allowRealApi"), "true")

	parsed, err := http.NewRequest(r.Method, targetURL, strings.NewReader(body))
	if err != nil {
		http.Error(w, `{"error":"invalid target URL"}`, http.StatusBadRequest)
		return
	}
	for key, values := range r.Header {
		if len(values) == 0 || shouldSkipProxyRequestHeader(key) {
			continue
		}
		parsed.Header.Set(key, values[0])
	}
	// Let Go's transport negotiate/decompress gzip so we always store/return plain text.
	parsed.Header.Del("Accept-Encoding")

	matchPath := endpoint
	if matchPath == "" && parsed.URL != nil {
		matchPath = parsed.URL.Path
	}
	if matchPath == "" {
		matchPath = "/"
	}

	forceAutoStub := false
	if isMockMode(h.Store) && !forceLive {
		if h.Store.Setting("virtualizationEnabled") == "false" {
			http.Error(w, `{"error":"virtualization is disabled"}`, http.StatusServiceUnavailable)
			return
		}
		stub, ok := h.Store.FindStubForAPI("", r.Method, matchPath, body)
		if !ok && parsed.URL != nil && matchPath != parsed.URL.Path {
			stub, ok = h.Store.FindStubForAPI("", r.Method, parsed.URL.Path, body)
		}
		if ok {
			h.writeStubResponse(w, r, body, targetURL, stub)
			return
		}
		// Toggle OFF (stubbed) with no matching stub: hit live API and auto-create stub.
		forceAutoStub = true
		w.Header().Set("X-No-Stub-Found", "true")
	}

	client := h.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	response, err := client.Do(parsed)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	responseBody, err := readDecodedBody(response)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to read upstream response"})
		return
	}
	respHeadersMap := make(map[string]string)
	for key, values := range response.Header {
		if len(values) == 0 || shouldSkipProxyResponseHeader(key) {
			continue
		}
		respHeadersMap[key] = values[0]
		w.Header().Set(key, values[0])
	}
	respHeadersJSON := encode(respHeadersMap)

	h.Store.SaveRequest(&RequestRecord{
		Method:          r.Method,
		URL:             targetURL,
		Endpoint:        matchPath,
		Headers:         encode(r.Header),
		Body:            body,
		Status:          response.StatusCode,
		Response:        string(responseBody),
		ResponseHeaders: respHeadersJSON,
		Source:          "live-api",
	})

	recSetting := strings.ToLower(h.Store.Setting("recordingMode"))
	shouldRecord := forceAutoStub || forceLive || (recSetting != "false" && recSetting != "off")
	if shouldRecord {
		h.recordLiveAsStub(r.Method, matchPath, body, r.Header.Get("X-Target-Host"), parsed, response.StatusCode, string(responseBody), respHeadersJSON)
	}

	w.Header().Set("X-Response-Source", "live-api")
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(responseBody)
}

func shouldSkipProxyRequestHeader(key string) bool {
	lk := strings.ToLower(key)
	switch lk {
	case "host", "connection", "content-length", "transfer-encoding", "accept-encoding",
		"if-none-match", "if-modified-since", "origin", "referer", "cookie",
		"x-forwarded-for", "x-forwarded-proto", "x-forwarded-host", "forwarded":
		return true
	}
	return strings.HasPrefix(lk, "sec-")
}

func shouldSkipProxyResponseHeader(key string) bool {
	lk := strings.ToLower(key)
	switch lk {
	case "content-encoding", "content-length", "transfer-encoding", "connection",
		"keep-alive", "proxy-authenticate", "proxy-authorization", "trailer", "upgrade",
		"access-control-allow-origin", "access-control-allow-credentials",
		"access-control-allow-methods", "access-control-allow-headers",
		"access-control-expose-headers", "access-control-max-age":
		return true
	}
	return false
}

func readDecodedBody(response *http.Response) ([]byte, error) {
	var reader io.Reader = response.Body
	encoding := strings.ToLower(response.Header.Get("Content-Encoding"))
	if strings.Contains(encoding, "gzip") {
		gzReader, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	} else if strings.Contains(encoding, "deflate") {
		flReader := flate.NewReader(response.Body)
		defer flReader.Close()
		reader = flReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	// Fallback: body still looks gzip-compressed (e.g. encoding header stripped incorrectly).
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b {
		if gzReader, gzErr := gzip.NewReader(bytes.NewReader(body)); gzErr == nil {
			defer gzReader.Close()
			if decoded, decErr := io.ReadAll(gzReader); decErr == nil {
				return decoded, nil
			}
		}
	}
	return body, nil
}

func (h *RuntimeHandler) recordLiveAsStub(method, matchPath, body, targetHost string, parsed *http.Request, status int, responseBody, respHeadersJSON string) {
	var existing *Stub
	for _, s := range h.Store.ListStubs() {
		if methodMatches(s.Method, method) && pathMatches(s.Endpoint, matchPath) {
			existing = s
			break
		}
	}
	if existing != nil {
		matchesActive := existing.ResponseStatus == status && existing.ResponseBody == responseBody
		existingVersions := h.Store.ListVersions(existing.ID)
		matchesVersion := false
		for _, v := range existingVersions {
			if v.ResponseStatus == status && v.ResponseBody == responseBody {
				matchesVersion = true
				break
			}
		}
		if !matchesActive && !matchesVersion {
			h.Store.SaveVersion(&StubVersion{
				StubID:          existing.ID,
				VersionTag:      fmt.Sprintf("Auto-recorded: %d", status),
				ResponseStatus:  status,
				ResponseBody:    responseBody,
				ResponseHeaders: respHeadersJSON,
				Active:          false,
			})
		}
		return
	}

	apiID := ""
	if api := identifyAPI(h.Store.ListAPIs(), method, matchPath); api != nil {
		apiID = api.ID
	}
	if targetHost == "" && parsed != nil && parsed.URL != nil {
		targetHost = parsed.URL.Scheme + "://" + parsed.URL.Host
	}
	h.Store.SaveStub(&Stub{
		APIID:           apiID,
		Name:            "Auto-recorded: " + matchPath,
		Method:          method,
		Endpoint:        matchPath,
		BaseURL:         targetHost,
		Environment:     "Dev",
		Category:        "Other",
		Description:     "Auto-recorded from live traffic",
		RequestMatcher:  body,
		ResponseStatus:  status,
		ResponseBody:    responseBody,
		ResponseHeaders: respHeadersJSON,
		Enabled:         true,
		Version:         "v1",
	})
}

func (h *RuntimeHandler) writeStubResponse(w http.ResponseWriter, r *http.Request, body string, requestURL string, stub *Stub) {
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

	// Record every stub-served request so it appears in Recorded Requests.
	if r != nil {
		endpoint := r.URL.Query().Get("endpoint")
		if endpoint == "" {
			endpoint = r.URL.Path
		}
		if requestURL == "" {
			requestURL = r.URL.String()
		}
		h.Store.SaveRequest(&RequestRecord{
			Method:          r.Method,
			URL:             requestURL,
			BaseURL:         stub.BaseURL,
			Endpoint:        endpoint,
			Headers:         encode(r.Header),
			Body:            body,
			Status:          stub.ResponseStatus,
			Response:        responseBody,
			ResponseHeaders: stub.ResponseHeaders,
			Source:          "stub",
			Category:        stub.Category,
			OwnerGroup:      stub.OwnerGroup,
		})
	}
}

type CaptureHandler struct {
	Store  Store
	Blobs  BlobStore
	Client *http.Client
}

type HealthCheckHandler struct{ Client *http.Client }

func (h *HealthCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid target URL"})
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
