package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// APITestFields is a test implementation of AirtableFields for API tests
type APITestFields struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func TestUpdateRecordsRequest(t *testing.T) {
	baseID := "app123abc"
	tableName := "TestTable"
	patchBody := bytes.NewBufferString(`{"records":[]}`)
	ctx := context.Background()

	request, err := UpdateRecordsRequest(baseID, tableName, patchBody, ctx)
	if err != nil {
		t.Fatalf("UpdateRecordsRequest() error: %v", err)
	}

	if request == nil {
		t.Fatal("Request should not be nil")
	}

	if request.Method != http.MethodPatch {
		t.Errorf("Method = %q, want %q", request.Method, http.MethodPatch)
	}

	expectedURL := "https://api.airtable.com/v0/app123abc/TestTable"
	if request.URL.String() != expectedURL {
		t.Errorf("URL = %q, want %q", request.URL.String(), expectedURL)
	}

	contentType := request.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestUpdateRecordsRequestNilContext(t *testing.T) {
	baseID := "app123"
	tableName := "Table"
	patchBody := bytes.NewBufferString(`{}`)

	request, err := UpdateRecordsRequest(baseID, tableName, patchBody, nil)
	if err != nil {
		t.Fatalf("UpdateRecordsRequest() with nil context error: %v", err)
	}

	if request == nil {
		t.Fatal("Request should not be nil")
	}
}

func TestEncodePostAsJSON(t *testing.T) {
	request := CreateRecordsRequest[APITestFields]{
		Records: []CreateRecordRequest[APITestFields]{
			{
				Fields: APITestFields{Name: "Test", Status: "Active"},
			},
		},
		Typecast: true,
	}

	buf, err := EncodePostAsJSON(request)
	if err != nil {
		t.Fatalf("EncodePostAsJSON() error: %v", err)
	}

	if buf == nil {
		t.Fatal("Buffer should not be nil")
	}

	// Verify JSON is valid
	var decoded CreateRecordsRequest[APITestFields]
	if err := json.NewDecoder(buf).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(decoded.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(decoded.Records))
	}

	if decoded.Records[0].Fields.Name != "Test" {
		t.Errorf("Fields.Name = %q, want %q", decoded.Records[0].Fields.Name, "Test")
	}
}

func TestEncodePostAsJSONEmpty(t *testing.T) {
	request := CreateRecordsRequest[APITestFields]{
		Records:  []CreateRecordRequest[APITestFields]{},
		Typecast: false,
	}

	buf, err := EncodePostAsJSON(request)
	if err != nil {
		t.Fatalf("EncodePostAsJSON() error: %v", err)
	}

	if buf == nil {
		t.Fatal("Buffer should not be nil")
	}

	if buf.Len() == 0 {
		t.Error("Buffer should not be empty")
	}
}

func TestEncodePatchAsJSON(t *testing.T) {
	request := PatchItemsRequest[APITestFields]{
		Records: []PatchItemRequest[APITestFields]{
			{
				Id:     "rec123",
				Fields: APITestFields{Name: "Updated", Status: "Inactive"},
			},
		},
		Typecast: true,
	}

	buf, err := EncodePatchAsJSON(request)
	if err != nil {
		t.Fatalf("EncodePatchAsJSON() error: %v", err)
	}

	if buf == nil {
		t.Fatal("Buffer should not be nil")
	}

	// Verify JSON is valid
	var decoded PatchItemsRequest[APITestFields]
	if err := json.NewDecoder(buf).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(decoded.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(decoded.Records))
	}

	if decoded.Records[0].Id != "rec123" {
		t.Errorf("Records[0].Id = %q, want %q", decoded.Records[0].Id, "rec123")
	}
}

func TestEncodePatchAsJSONEmpty(t *testing.T) {
	request := PatchItemsRequest[APITestFields]{
		Records:  []PatchItemRequest[APITestFields]{},
		Typecast: false,
	}

	buf, err := EncodePatchAsJSON(request)
	if err != nil {
		t.Fatalf("EncodePatchAsJSON() error: %v", err)
	}

	if buf == nil {
		t.Fatal("Buffer should not be nil")
	}
}

func TestEncodePatchAsJSONMultipleRecords(t *testing.T) {
	request := PatchItemsRequest[APITestFields]{
		Records: []PatchItemRequest[APITestFields]{
			{Id: "rec1", Fields: APITestFields{Name: "First", Status: "A"}},
			{Id: "rec2", Fields: APITestFields{Name: "Second", Status: "B"}},
			{Id: "rec3", Fields: APITestFields{Name: "Third", Status: "C"}},
		},
		Typecast: true,
	}

	buf, err := EncodePatchAsJSON(request)
	if err != nil {
		t.Fatalf("EncodePatchAsJSON() error: %v", err)
	}

	var decoded PatchItemsRequest[APITestFields]
	if err := json.NewDecoder(buf).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(decoded.Records) != 3 {
		t.Errorf("Records length = %d, want 3", len(decoded.Records))
	}
}

func TestHandleErrorResponseTooManyRequests(t *testing.T) {
	err := handleErrorResponse(http.StatusTooManyRequests, []byte{})
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("Error = %q, should contain 'rate limited'", err.Error())
	}
}

func TestHandleErrorResponseUnauthorized(t *testing.T) {
	err := handleErrorResponse(http.StatusUnauthorized, []byte{})
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Error = %q, should contain 'unauthorized'", err.Error())
	}
}

func TestHandleErrorResponseForbidden(t *testing.T) {
	body := []byte(`{"error":{"type":"FORBIDDEN","message":"Access denied"}}`)
	err := handleErrorResponse(http.StatusForbidden, body)
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Error = %q, should contain 'forbidden'", err.Error())
	}
}

func TestHandleErrorResponseNotFound(t *testing.T) {
	body := []byte(`{"error":{"type":"NOT_FOUND","message":"Record not found"}}`)
	err := handleErrorResponse(http.StatusNotFound, body)
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error = %q, should contain 'not found'", err.Error())
	}
}

func TestHandleErrorResponseRequestEntityTooLarge(t *testing.T) {
	err := handleErrorResponse(http.StatusRequestEntityTooLarge, []byte{})
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("Error = %q, should contain 'too large'", err.Error())
	}
}

func TestHandleErrorResponseUnprocessableEntity(t *testing.T) {
	body := []byte(`{"error":{"type":"INVALID_REQUEST","message":"Invalid field"}}`)
	err := handleErrorResponse(http.StatusUnprocessableEntity, body)
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "unprocessable") {
		t.Errorf("Error = %q, should contain 'unprocessable'", err.Error())
	}
}

func TestHandleErrorResponseDefault(t *testing.T) {
	body := []byte(`some error body`)
	err := handleErrorResponse(http.StatusInternalServerError, body)
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Error = %q, should contain status code '500'", err.Error())
	}
}

func TestGetErrorWithErrorResponse(t *testing.T) {
	body := []byte(`{"error":{"type":"TEST_ERROR","message":"Test error message"}}`)
	err := getError(body, "test context")
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "TEST_ERROR") {
		t.Errorf("Error = %q, should contain 'TEST_ERROR'", err.Error())
	}

	if !strings.Contains(err.Error(), "Test error message") {
		t.Errorf("Error = %q, should contain 'Test error message'", err.Error())
	}
}

func TestGetErrorWithSimpleErrorResponse(t *testing.T) {
	body := []byte(`{"error":"Simple error string"}`)
	err := getError(body, "test context")
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "Simple error string") {
		t.Errorf("Error = %q, should contain 'Simple error string'", err.Error())
	}
}

func TestGetErrorWithInvalidJSON(t *testing.T) {
	body := []byte(`not valid json`)
	err := getError(body, "test context")
	if err == nil {
		t.Fatal("Error should not be nil for invalid JSON")
	}
}

func TestGetErrorFromErrorResponseWithMessage(t *testing.T) {
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Type:    "ERROR_TYPE",
			Message: "Error message",
		},
	}

	err := getErrorFromErrorResponse(errorResponse, "context")
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "ERROR_TYPE") {
		t.Errorf("Error = %q, should contain 'ERROR_TYPE'", err.Error())
	}

	if !strings.Contains(err.Error(), "Error message") {
		t.Errorf("Error = %q, should contain 'Error message'", err.Error())
	}
}

func TestGetErrorFromErrorResponseWithoutMessage(t *testing.T) {
	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Type:    "ERROR_TYPE",
			Message: "",
		},
	}

	err := getErrorFromErrorResponse(errorResponse, "context")
	if err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(err.Error(), "ERROR_TYPE") {
		t.Errorf("Error = %q, should contain 'ERROR_TYPE'", err.Error())
	}
}

func TestListRecordsRequestBasic(t *testing.T) {
	baseID := "appXYZ"
	tableName := "MyTable"
	viewName := "Grid view"
	maxRecords := 50
	offset := "first_call"

	request, err := listRecordsRequest(baseID, tableName, viewName, nil, maxRecords, offset)
	if err != nil {
		t.Fatalf("listRecordsRequest() error: %v", err)
	}

	if request.Method != http.MethodGet {
		t.Errorf("Method = %q, want %q", request.Method, http.MethodGet)
	}

	// Check URL path
	if !strings.Contains(request.URL.String(), baseID) {
		t.Errorf("URL should contain baseID")
	}

	if !strings.Contains(request.URL.String(), tableName) {
		t.Errorf("URL should contain tableName")
	}

	// Check query params
	query := request.URL.Query()
	if query.Get("maxRecords") != "50" {
		t.Errorf("maxRecords = %q, want %q", query.Get("maxRecords"), "50")
	}

	if query.Get("view") != viewName {
		t.Errorf("view = %q, want %q", query.Get("view"), viewName)
	}

	// offset should not be included for first_call
	if query.Get("offset") != "" {
		t.Errorf("offset should not be set for first_call")
	}
}

func TestListRecordsRequestWithOffset(t *testing.T) {
	baseID := "appXYZ"
	tableName := "MyTable"
	viewName := "Grid view"
	maxRecords := 100
	offset := "next_page_token_123"

	request, err := listRecordsRequest(baseID, tableName, viewName, nil, maxRecords, offset)
	if err != nil {
		t.Fatalf("listRecordsRequest() error: %v", err)
	}

	query := request.URL.Query()
	if query.Get("offset") != offset {
		t.Errorf("offset = %q, want %q", query.Get("offset"), offset)
	}
}

func TestListRecordsRequestDefaultMaxRecords(t *testing.T) {
	baseID := "appXYZ"
	tableName := "MyTable"
	viewName := "Grid view"
	maxRecords := 0 // Should default to 100
	offset := "first_call"

	request, err := listRecordsRequest(baseID, tableName, viewName, nil, maxRecords, offset)
	if err != nil {
		t.Fatalf("listRecordsRequest() error: %v", err)
	}

	query := request.URL.Query()
	if query.Get("maxRecords") != "100" {
		t.Errorf("maxRecords = %q, want %q (default)", query.Get("maxRecords"), "100")
	}
}

func TestListRecordsRequestWithContext(t *testing.T) {
	baseID := "appXYZ"
	tableName := "MyTable"
	viewName := "Grid view"
	ctx := context.Background()

	request, err := listRecordsRequest(baseID, tableName, viewName, ctx, 50, "first_call")
	if err != nil {
		t.Fatalf("listRecordsRequest() error: %v", err)
	}

	if request.Context() != ctx {
		t.Error("Request context should match provided context")
	}
}

// Test constants
func TestAPIConstants(t *testing.T) {
	if defaultMaxRecords != 100 {
		t.Errorf("defaultMaxRecords = %d, want 100", defaultMaxRecords)
	}

	if contentTypeJSON != "application/json" {
		t.Errorf("contentTypeJSON = %q, want %q", contentTypeJSON, "application/json")
	}

	if apiBasePath != "https://api.airtable.com/v0" {
		t.Errorf("apiBasePath = %q, want %q", apiBasePath, "https://api.airtable.com/v0")
	}

	if defaultOffset != "first_call" {
		t.Errorf("defaultOffset = %q, want %q", defaultOffset, "first_call")
	}
}

// Benchmark tests
func BenchmarkEncodePostAsJSON(b *testing.B) {
	request := CreateRecordsRequest[APITestFields]{
		Records: []CreateRecordRequest[APITestFields]{
			{Fields: APITestFields{Name: "Test", Status: "Active"}},
		},
		Typecast: true,
	}

	for i := 0; i < b.N; i++ {
		_, _ = EncodePostAsJSON(request)
	}
}

func BenchmarkEncodePatchAsJSON(b *testing.B) {
	request := PatchItemsRequest[APITestFields]{
		Records: []PatchItemRequest[APITestFields]{
			{Id: "rec123", Fields: APITestFields{Name: "Test", Status: "Active"}},
		},
		Typecast: true,
	}

	for i := 0; i < b.N; i++ {
		_, _ = EncodePatchAsJSON(request)
	}
}

func BenchmarkHandleErrorResponse(b *testing.B) {
	body := []byte(`{"error":{"type":"TEST_ERROR","message":"Test error message"}}`)

	for i := 0; i < b.N; i++ {
		_ = handleErrorResponse(http.StatusForbidden, body)
	}
}

// Tests using httptest for HTTP client functions

func TestListRecords(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		contentType    string
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful list records",
			serverResponse: `{
				"records": [
					{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "First", "status": "Active"}},
					{"id": "rec2", "createdTime": "2023-01-02T00:00:00Z", "fields": {"name": "Second", "status": "Inactive"}}
				],
				"offset": ""
			}`,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			wantErr:     false,
			wantCount:   2,
		},
		{
			name:           "empty records",
			serverResponse: `{"records": [], "offset": ""}`,
			statusCode:     http.StatusOK,
			contentType:    "application/json",
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "error response - unauthorized",
			serverResponse: `{"error":{"type":"UNAUTHORIZED","message":"Invalid token"}}`,
			statusCode:     http.StatusUnauthorized,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "error response - rate limited",
			serverResponse: `{}`,
			statusCode:     http.StatusTooManyRequests,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "wrong content type",
			serverResponse: `{"records": []}`,
			statusCode:     http.StatusOK,
			contentType:    "text/html",
			wantErr:        true,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Override the API base path to use test server
			SetAPIBasePath(server.URL)
			defer ResetAPIBasePath()

			records, err := ListRecords[APITestFields](server.Client(), "", "", "", nil, 100)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(records) != tt.wantCount {
				t.Errorf("ListRecords() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestListRecordsPagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++

		if callCount == 1 {
			// First page with offset
			_, _ = w.Write([]byte(`{
				"records": [
					{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "First", "status": "Active"}}
				],
				"offset": "page2token"
			}`))
		} else {
			// Second page without offset (last page)
			_, _ = w.Write([]byte(`{
				"records": [
					{"id": "rec2", "createdTime": "2023-01-02T00:00:00Z", "fields": {"name": "Second", "status": "Inactive"}}
				],
				"offset": ""
			}`))
		}
	}))
	defer server.Close()

	// Override the API base path to use test server
	SetAPIBasePath(server.URL)
	defer ResetAPIBasePath()

	records, err := ListRecords[APITestFields](server.Client(), "", "", "", nil, 100)
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}

	if len(records) != 2 {
		t.Errorf("ListRecords() returned %d records, want 2", len(records))
	}

	if callCount != 2 {
		t.Errorf("Server called %d times, want 2", callCount)
	}
}

func TestListRecordsWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"records": [{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "Test", "status": "Active"}}], "offset": ""}`))
	}))
	defer server.Close()

	SetAPIBasePath(server.URL)
	defer ResetAPIBasePath()

	ctx := context.Background()
	records, err := ListRecords[APITestFields](server.Client(), "baseID", "tableName", "viewName", ctx, 50)

	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}

	if len(records) != 1 {
		t.Errorf("ListRecords() returned %d records, want 1", len(records))
	}
}

func TestListRecordsDefaultMaxRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify maxRecords query parameter defaults to 100
		maxRecords := r.URL.Query().Get("maxRecords")
		if maxRecords != "100" {
			t.Errorf("maxRecords = %q, want %q", maxRecords, "100")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"records": [], "offset": ""}`))
	}))
	defer server.Close()

	SetAPIBasePath(server.URL)
	defer ResetAPIBasePath()

	_, err := ListRecords[APITestFields](server.Client(), "baseID", "tableName", "viewName", nil, 0)
	if err != nil {
		t.Fatalf("ListRecords() error = %v", err)
	}
}

func TestUpdateRecords(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		contentType    string
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful update",
			serverResponse: `{
				"records": [
					{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "Updated", "status": "Active"}}
				]
			}`,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name: "update multiple records",
			serverResponse: `{
				"records": [
					{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "First", "status": "A"}},
					{"id": "rec2", "createdTime": "2023-01-02T00:00:00Z", "fields": {"name": "Second", "status": "B"}}
				]
			}`,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			wantErr:     false,
			wantCount:   2,
		},
		{
			name:           "error response - forbidden",
			serverResponse: `{"error":{"type":"FORBIDDEN","message":"Access denied"}}`,
			statusCode:     http.StatusForbidden,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "error response - not found",
			serverResponse: `{"error":{"type":"NOT_FOUND","message":"Record not found"}}`,
			statusCode:     http.StatusNotFound,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "wrong content type",
			serverResponse: `{"records": []}`,
			statusCode:     http.StatusOK,
			contentType:    "text/plain",
			wantErr:        true,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			patchBody := bytes.NewBufferString(`{"records":[{"id":"rec1","fields":{"name":"Updated"}}]}`)
			request, _ := http.NewRequest(http.MethodPatch, server.URL, patchBody)
			request.Header.Set("Content-Type", "application/json")

			records, err := UpdateRecords[APITestFields](server.Client(), request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(records) != tt.wantCount {
				t.Errorf("UpdateRecords() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestCreateRecord(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		contentType    string
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful create",
			serverResponse: `{
				"records": [
					{"id": "recNew", "createdTime": "2023-06-15T00:00:00Z", "fields": {"name": "New Record", "status": "Active"}}
				]
			}`,
			statusCode:  http.StatusOK,
			contentType: "application/json",
			wantErr:     false,
			wantCount:   1,
		},
		{
			name:           "error response - unprocessable entity",
			serverResponse: `{"error":{"type":"INVALID_REQUEST","message":"Invalid field"}}`,
			statusCode:     http.StatusUnprocessableEntity,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "error response - request too large",
			serverResponse: `{}`,
			statusCode:     http.StatusRequestEntityTooLarge,
			contentType:    "application/json",
			wantErr:        true,
			wantCount:      0,
		},
		{
			name:           "wrong content type",
			serverResponse: `{"records": []}`,
			statusCode:     http.StatusOK,
			contentType:    "text/xml",
			wantErr:        true,
			wantCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Verify content type header
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Override the API base path to use test server
			SetAPIBasePath(server.URL)
			defer ResetAPIBasePath()

			record := &APITestFields{Name: "New Record", Status: "Active"}
			records, err := CreateRecord[APITestFields, APITestFields](server.Client(), record, "", "", nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(records) != tt.wantCount {
				t.Errorf("CreateRecord() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestCreateRecordWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"records": [{"id": "rec1", "createdTime": "2023-01-01T00:00:00Z", "fields": {"name": "Test", "status": "Active"}}]}`))
	}))
	defer server.Close()

	SetAPIBasePath(server.URL)
	defer ResetAPIBasePath()

	ctx := context.Background()
	record := &APITestFields{Name: "Test", Status: "Active"}
	records, err := CreateRecord[APITestFields, APITestFields](server.Client(), record, "baseID", "tableName", ctx)

	if err != nil {
		t.Fatalf("CreateRecord() error = %v", err)
	}

	if len(records) != 1 {
		t.Errorf("CreateRecord() returned %d records, want 1", len(records))
	}
}

func TestSetAndResetAPIBasePath(t *testing.T) {
	// Verify initial state
	if apiBasePath != defaultAPIBasePath {
		t.Errorf("Initial apiBasePath = %q, want %q", apiBasePath, defaultAPIBasePath)
	}

	// Set a custom path
	customPath := "https://custom.api.com/v1"
	SetAPIBasePath(customPath)

	if apiBasePath != customPath {
		t.Errorf("After SetAPIBasePath, apiBasePath = %q, want %q", apiBasePath, customPath)
	}

	// Reset to default
	ResetAPIBasePath()

	if apiBasePath != defaultAPIBasePath {
		t.Errorf("After ResetAPIBasePath, apiBasePath = %q, want %q", apiBasePath, defaultAPIBasePath)
	}
}

func TestHTTPDoerInterface(t *testing.T) {
	// Verify that *http.Client satisfies HTTPDoer interface
	var _ HTTPDoer = &http.Client{}
}
