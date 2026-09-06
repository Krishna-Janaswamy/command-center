package virtualization

import "time"

type API struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Method      string    `json:"method"`
	Endpoint    string    `json:"endpoint"`
	BaseURL     string    `json:"baseUrl"`
	Environment string    `json:"environment"`
	Category    string    `json:"category"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Stub struct {
	ID              string `json:"id"`
	APIID           string `json:"apiId"`
	Name            string `json:"name"`
	Method          string `json:"method"`
	Endpoint        string `json:"endpoint"`
	BaseURL         string `json:"baseUrl"`
	RequestMatcher  string `json:"requestMatcher"`
	ResponseStatus  int    `json:"responseStatus"`
	ResponseBody    string `json:"responseBody"`
	ResponseHeaders string `json:"responseHeaders"`
	DelayMS         int    `json:"delay"`
	Enabled         bool   `json:"enabled"`
	Version         string `json:"version"`
}

type StubVersion struct {
	ID              string `json:"versionId"`
	StubID          string `json:"stubId"`
	Version         string `json:"version"`
	ResponseStatus  int    `json:"responseStatus"`
	ResponseBody    string `json:"responseBody"`
	ResponseHeaders string `json:"responseHeaders"`
	Active          bool   `json:"active"`
}

type RequestRecord struct {
	ID              string    `json:"id"`
	Method          string    `json:"method"`
	URL             string    `json:"url"`
	Endpoint        string    `json:"endpoint"`
	Headers         string    `json:"headers"`
	Body            string    `json:"body"`
	BodyS3Key       string    `json:"bodyS3Key,omitempty"`
	Status          int       `json:"status"`
	Response        string    `json:"response"`
	ResponseS3Key   string    `json:"responseS3Key,omitempty"`
	ResponseHeaders string    `json:"responseHeaders"`
	Source          string    `json:"source"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CaptureRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
type CaptureResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
