package httphelper

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConstants(t *testing.T) {
	// Verify header constants are set correctly
	if HeaderHost != "Host" {
		t.Errorf("HeaderHost = %q, want %q", HeaderHost, "Host")
	}
	if HeaderXForwardedFor != "X-Forwarded-For" {
		t.Errorf("HeaderXForwardedFor = %q, want %q", HeaderXForwardedFor, "X-Forwarded-For")
	}
	if HeaderXForwardedHost != "X-Forwarded-Host" {
		t.Errorf("HeaderXForwardedHost = %q, want %q", HeaderXForwardedHost, "X-Forwarded-Host")
	}
}

func TestSetBearerTokenHeader(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantAuth string
	}{
		{
			name:     "simple token",
			token:    "abc123",
			wantAuth: "Bearer abc123",
		},
		{
			name:     "empty token",
			token:    "",
			wantAuth: "Bearer ",
		},
		{
			name:     "token with special characters",
			token:    "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			wantAuth: "Bearer ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:     "JWT-like token",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			wantAuth: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			SetBearerTokenHeader(req, tt.token)

			got := req.Header.Get("Authorization")
			if got != tt.wantAuth {
				t.Errorf("SetBearerTokenHeader() Authorization = %q, want %q", got, tt.wantAuth)
			}
		})
	}
}

func TestHasContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		checkType   string
		want        bool
	}{
		{
			name:        "exact match",
			contentType: "application/json",
			checkType:   "application/json",
			want:        true,
		},
		{
			name:        "with charset",
			contentType: "application/json; charset=utf-8",
			checkType:   "application/json",
			want:        true,
		},
		{
			name:        "with multiple params",
			contentType: "text/html; charset=utf-8; boundary=something",
			checkType:   "text/html",
			want:        true,
		},
		{
			name:        "no match",
			contentType: "application/json",
			checkType:   "text/html",
			want:        false,
		},
		{
			name:        "empty content type",
			contentType: "",
			checkType:   "application/json",
			want:        false,
		},
		{
			name:        "partial match is not a match",
			contentType: "application/json-patch+json",
			checkType:   "application/json",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{
					"Content-Type": []string{tt.contentType},
				},
			}
			got := HasContentType(resp, tt.checkType)
			if got != tt.want {
				t.Errorf("HasContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSuccessResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 edge case", 299, true},
		{"300 Multiple Choices", 300, false},
		{"301 Moved Permanently", 301, false},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"404 Not Found", 404, false},
		{"500 Internal Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode}
			got := IsSuccessResponse(resp)
			if got != tt.want {
				t.Errorf("IsSuccessResponse() with status %d = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestGetHeaders(t *testing.T) {
	tests := []struct {
		name      string
		keyValues map[string]string
		wantLen   int
	}{
		{
			name:      "empty map",
			keyValues: map[string]string{},
			wantLen:   0,
		},
		{
			name: "single header",
			keyValues: map[string]string{
				"Content-Type": "application/json",
			},
			wantLen: 1,
		},
		{
			name: "multiple headers",
			keyValues: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
				"Accept":        "application/json",
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := GetHeaders(tt.keyValues)
			if len(headers) != tt.wantLen {
				t.Errorf("GetHeaders() returned %d headers, want %d", len(headers), tt.wantLen)
			}
			for key, value := range tt.keyValues {
				if headers.Get(key) != value {
					t.Errorf("GetHeaders() header %q = %q, want %q", key, headers.Get(key), value)
				}
			}
		})
	}
}

func TestGetURL(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		params   map[string]string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "simple path no params",
			path:     "http://example.com/api",
			params:   nil,
			wantPath: "http://example.com/api",
			wantErr:  false,
		},
		{
			name:     "with single param",
			path:     "http://example.com/api",
			params:   map[string]string{"key": "value"},
			wantPath: "http://example.com/api?key=value",
			wantErr:  false,
		},
		{
			name:   "with multiple params",
			path:   "http://example.com/api",
			params: map[string]string{"key1": "value1", "key2": "value2"},
			// Note: map iteration order is not guaranteed, so we check params separately
			wantErr: false,
		},
		{
			name:    "path with existing query params",
			path:    "http://example.com/api?existing=param",
			params:  map[string]string{"new": "param"},
			wantErr: false,
		},
		{
			name:    "param with special characters",
			path:    "http://example.com/api",
			params:  map[string]string{"key": "value with spaces"},
			wantErr: false,
		},
		{
			name:     "empty params map",
			path:     "http://example.com/api",
			params:   map[string]string{},
			wantPath: "http://example.com/api",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetURL(tt.path, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Verify all params are present in the URL
			for key, value := range tt.params {
				if got.Query().Get(key) != value {
					t.Errorf("GetURL() query param %q = %q, want %q", key, got.Query().Get(key), value)
				}
			}
		})
	}
}

func TestGetURLWithInvalidPath(t *testing.T) {
	// Test with invalid URL that should cause parse error
	_, err := GetURL("://invalid-url", nil)
	if err == nil {
		t.Error("GetURL() with invalid URL should return error")
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		params     map[string]string
		headers    map[string]string
		body       io.Reader
		wantErr    bool
		wantMethod string
	}{
		{
			name:       "GET request",
			method:     "GET",
			path:       "http://example.com/api",
			params:     nil,
			headers:    nil,
			body:       nil,
			wantErr:    false,
			wantMethod: "GET",
		},
		{
			name:       "POST request with body",
			method:     "POST",
			path:       "http://example.com/api",
			params:     nil,
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"key": "value"}`),
			wantErr:    false,
			wantMethod: "POST",
		},
		{
			name:       "request with params and headers",
			method:     "GET",
			path:       "http://example.com/api",
			params:     map[string]string{"page": "1"},
			headers:    map[string]string{"Authorization": "Bearer token"},
			body:       nil,
			wantErr:    false,
			wantMethod: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewRequest(tt.method, tt.path, tt.params, tt.headers, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if req.Method != tt.wantMethod {
				t.Errorf("NewRequest() method = %q, want %q", req.Method, tt.wantMethod)
			}

			// Verify headers
			for key, value := range tt.headers {
				if req.Header.Get(key) != value {
					t.Errorf("NewRequest() header %q = %q, want %q", key, req.Header.Get(key), value)
				}
			}

			// Verify params
			for key, value := range tt.params {
				if req.URL.Query().Get(key) != value {
					t.Errorf("NewRequest() query param %q = %q, want %q", key, req.URL.Query().Get(key), value)
				}
			}
		})
	}
}

func TestNewRequestWithInvalidURL(t *testing.T) {
	_, err := NewRequest("GET", "://invalid", nil, nil, nil)
	if err == nil {
		t.Error("NewRequest() with invalid URL should return error")
	}
}

func TestGetBaseURL(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		xForwardedHost string
		wantBaseURL    string
	}{
		{
			name:        "simple localhost",
			host:        "localhost",
			wantBaseURL: "http://localhost",
		},
		{
			name:        "localhost with port 8080",
			host:        "localhost:8080",
			wantBaseURL: "http://localhost:8080",
		},
		{
			name:        "localhost with port 80 (omitted)",
			host:        "localhost:80",
			wantBaseURL: "http://localhost",
		},
		{
			name:           "proxied request with X-Forwarded-Host",
			host:           "localhost:8080",
			xForwardedHost: "example.com",
			wantBaseURL:    "https://example.com",
		},
		{
			name:           "proxied request with X-Forwarded-Host and port",
			host:           "localhost:8080",
			xForwardedHost: "example.com:443",
			wantBaseURL:    "https://example.com",
		},
		{
			name:        "domain without port",
			host:        "example.com",
			wantBaseURL: "http://example.com",
		},
		{
			name:        "empty host defaults to localhost",
			host:        "",
			wantBaseURL: "http://localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tt.host
			if tt.xForwardedHost != "" {
				req.Header.Set(HeaderXForwardedHost, tt.xForwardedHost)
			}

			got := GetBaseURL(req)
			if got != tt.wantBaseURL {
				t.Errorf("GetBaseURL() = %q, want %q", got, tt.wantBaseURL)
			}
		})
	}
}

func TestHasContentTypeInternal(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		contentType string
		want        bool
	}{
		{
			name:        "exact match",
			header:      "application/json",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "with charset - spaces around semicolon",
			header:      "application/json ; charset=utf-8",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "empty header",
			header:      "",
			contentType: "application/json",
			want:        false,
		},
		{
			name:        "empty content type check",
			header:      "application/json",
			contentType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasContentType(tt.header, tt.contentType)
			if got != tt.want {
				t.Errorf("hasContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHostWithoutPort(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "host with port",
			host: "localhost:8080",
			want: "localhost",
		},
		{
			name: "host without port",
			host: "localhost",
			want: "localhost",
		},
		{
			name: "domain with port",
			host: "example.com:443",
			want: "example.com",
		},
		{
			name: "IPv4 with port",
			host: "127.0.0.1:8080",
			want: "127.0.0.1",
		},
		{
			name: "IPv6 with port",
			host: "[::1]:8080",
			want: "::1",
		},
		{
			name: "empty string",
			host: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getHostWithoutPort(tt.host)
			if got != tt.want {
				t.Errorf("getHostWithoutPort(%q) = %q, want %q", tt.host, got, tt.want)
			}
		})
	}
}

func TestGetDomainFromHostHeaders(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		xForwardedHost string
		want           string
	}{
		{
			name: "from Host header",
			host: "example.com:8080",
			want: "example.com",
		},
		{
			name:           "X-Forwarded-Host takes precedence",
			host:           "localhost:8080",
			xForwardedHost: "example.com",
			want:           "example.com",
		},
		{
			name: "empty host defaults to localhost",
			host: "",
			want: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tt.host
			if tt.xForwardedHost != "" {
				req.Header.Set(HeaderXForwardedHost, tt.xForwardedHost)
			}

			got := getDomainFromHostHeaders(req)
			if got != tt.want {
				t.Errorf("getDomainFromHostHeaders() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetProtocolFromHostHeaders(t *testing.T) {
	tests := []struct {
		name           string
		xForwardedHost string
		want           string
	}{
		{
			name: "no X-Forwarded-Host returns http",
			want: "http",
		},
		{
			name:           "with X-Forwarded-Host returns https",
			xForwardedHost: "example.com",
			want:           "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.xForwardedHost != "" {
				req.Header.Set(HeaderXForwardedHost, tt.xForwardedHost)
			}

			got := getProtocolFromHostHeaders(req)
			if got != tt.want {
				t.Errorf("getProtocolFromHostHeaders() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetPortFromHostHeaders(t *testing.T) {
	tests := []struct {
		name           string
		host           string
		xForwardedHost string
		want           string
	}{
		{
			name: "port 8080",
			host: "localhost:8080",
			want: ":8080",
		},
		{
			name: "port 80 returns empty",
			host: "localhost:80",
			want: "",
		},
		{
			name: "no port returns empty",
			host: "localhost",
			want: "",
		},
		{
			name:           "proxied request returns empty",
			host:           "localhost:8080",
			xForwardedHost: "example.com",
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.Host = tt.host
			if tt.xForwardedHost != "" {
				req.Header.Set(HeaderXForwardedHost, tt.xForwardedHost)
			}

			got := getPortFromHostHeaders(req)
			if got != tt.want {
				t.Errorf("getPortFromHostHeaders() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDumpResponse(t *testing.T) {
	// Create a response with a body
	body := "test response body"
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(body)),
	}

	// Capture stdout - this is a simple test to ensure no panic
	// In production, you'd use a more sophisticated approach
	err := DumpResponse(resp)
	if err != nil {
		t.Errorf("DumpResponse() returned error: %v", err)
	}
}

func TestDumpResponseWithEmptyBody(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader([]byte{})),
	}

	err := DumpResponse(resp)
	if err != nil {
		t.Errorf("DumpResponse() with empty body returned error: %v", err)
	}
}

// Benchmark tests
func BenchmarkSetBearerTokenHeader(b *testing.B) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	token := "test_token_12345"

	for i := 0; i < b.N; i++ {
		SetBearerTokenHeader(req, token)
	}
}

func BenchmarkHasContentType(b *testing.B) {
	resp := &http.Response{
		Header: http.Header{
			"Content-Type": []string{"application/json; charset=utf-8"},
		},
	}

	for i := 0; i < b.N; i++ {
		HasContentType(resp, "application/json")
	}
}

func BenchmarkGetHeaders(b *testing.B) {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer token",
		"Accept":        "application/json",
	}

	for i := 0; i < b.N; i++ {
		GetHeaders(headers)
	}
}

func BenchmarkGetURL(b *testing.B) {
	params := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	for i := 0; i < b.N; i++ {
		_, _ = GetURL("http://example.com/api", params)
	}
}

func BenchmarkGetBaseURL(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Host = "localhost:8080"

	for i := 0; i < b.N; i++ {
		GetBaseURL(req)
	}
}

func BenchmarkNewRequest(b *testing.B) {
	params := map[string]string{"key": "value"}
	headers := map[string]string{"Authorization": "Bearer token"}

	for i := 0; i < b.N; i++ {
		_, _ = NewRequest("GET", "http://example.com/api", params, headers, nil)
	}
}
