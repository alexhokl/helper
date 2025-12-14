package airtable

import (
	"encoding/json"
	"testing"
	"time"
)

// TestFields is a test implementation of AirtableFields
type TestFields struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestAirtableRecordJSONMarshaling(t *testing.T) {
	record := AirtableRecord[TestFields]{
		Id:          "rec123abc",
		CreatedTime: time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC),
		Fields: TestFields{
			Name:  "Test",
			Value: 42,
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("Failed to marshal AirtableRecord: %v", err)
	}

	// Unmarshal back
	var decoded AirtableRecord[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AirtableRecord: %v", err)
	}

	if decoded.Id != record.Id {
		t.Errorf("Id = %q, want %q", decoded.Id, record.Id)
	}

	if !decoded.CreatedTime.Equal(record.CreatedTime) {
		t.Errorf("CreatedTime = %v, want %v", decoded.CreatedTime, record.CreatedTime)
	}

	if decoded.Fields.Name != record.Fields.Name {
		t.Errorf("Fields.Name = %q, want %q", decoded.Fields.Name, record.Fields.Name)
	}

	if decoded.Fields.Value != record.Fields.Value {
		t.Errorf("Fields.Value = %d, want %d", decoded.Fields.Value, record.Fields.Value)
	}
}

func TestAirtableRecordsJSONMarshaling(t *testing.T) {
	records := AirtableRecords[TestFields]{
		Records: []AirtableRecord[TestFields]{
			{
				Id:          "rec1",
				CreatedTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Fields:      TestFields{Name: "First", Value: 1},
			},
			{
				Id:          "rec2",
				CreatedTime: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				Fields:      TestFields{Name: "Second", Value: 2},
			},
		},
		Offset: "next_page_token",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("Failed to marshal AirtableRecords: %v", err)
	}

	// Unmarshal back
	var decoded AirtableRecords[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AirtableRecords: %v", err)
	}

	if len(decoded.Records) != 2 {
		t.Errorf("Records length = %d, want 2", len(decoded.Records))
	}

	if decoded.Offset != records.Offset {
		t.Errorf("Offset = %q, want %q", decoded.Offset, records.Offset)
	}
}

func TestAirtableRecordsEmptyOffset(t *testing.T) {
	records := AirtableRecords[TestFields]{
		Records: []AirtableRecord[TestFields]{},
		Offset:  "",
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		t.Fatalf("Failed to marshal AirtableRecords: %v", err)
	}

	var decoded AirtableRecords[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal AirtableRecords: %v", err)
	}

	if decoded.Offset != "" {
		t.Errorf("Offset = %q, want empty string", decoded.Offset)
	}
}

func TestPatchItemsRequestJSONMarshaling(t *testing.T) {
	request := PatchItemsRequest[TestFields]{
		Records: []PatchItemRequest[TestFields]{
			{
				Id:     "rec1",
				Fields: TestFields{Name: "Updated", Value: 100},
			},
		},
		Typecast: true,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal PatchItemsRequest: %v", err)
	}

	var decoded PatchItemsRequest[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PatchItemsRequest: %v", err)
	}

	if len(decoded.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(decoded.Records))
	}

	if decoded.Typecast != true {
		t.Errorf("Typecast = %v, want true", decoded.Typecast)
	}

	if decoded.Records[0].Id != "rec1" {
		t.Errorf("Records[0].Id = %q, want %q", decoded.Records[0].Id, "rec1")
	}
}

func TestPatchItemRequestJSONMarshaling(t *testing.T) {
	request := PatchItemRequest[TestFields]{
		Id:     "rec123",
		Fields: TestFields{Name: "Test", Value: 50},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal PatchItemRequest: %v", err)
	}

	var decoded PatchItemRequest[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PatchItemRequest: %v", err)
	}

	if decoded.Id != request.Id {
		t.Errorf("Id = %q, want %q", decoded.Id, request.Id)
	}

	if decoded.Fields.Name != request.Fields.Name {
		t.Errorf("Fields.Name = %q, want %q", decoded.Fields.Name, request.Fields.Name)
	}
}

func TestCreateRecordsRequestJSONMarshaling(t *testing.T) {
	request := CreateRecordsRequest[TestFields]{
		Records: []CreateRecordRequest[TestFields]{
			{
				Fields: TestFields{Name: "New Record", Value: 999},
			},
		},
		Typecast: true,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal CreateRecordsRequest: %v", err)
	}

	var decoded CreateRecordsRequest[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CreateRecordsRequest: %v", err)
	}

	if len(decoded.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(decoded.Records))
	}

	if decoded.Typecast != true {
		t.Errorf("Typecast = %v, want true", decoded.Typecast)
	}
}

func TestCreateRecordRequestJSONMarshaling(t *testing.T) {
	request := CreateRecordRequest[TestFields]{
		Fields: TestFields{Name: "Created", Value: 123},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal CreateRecordRequest: %v", err)
	}

	var decoded CreateRecordRequest[TestFields]
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal CreateRecordRequest: %v", err)
	}

	if decoded.Fields.Name != request.Fields.Name {
		t.Errorf("Fields.Name = %q, want %q", decoded.Fields.Name, request.Fields.Name)
	}

	if decoded.Fields.Value != request.Fields.Value {
		t.Errorf("Fields.Value = %d, want %d", decoded.Fields.Value, request.Fields.Value)
	}
}

func TestErrorResponseJSONMarshaling(t *testing.T) {
	errorResp := ErrorResponse{
		Error: ErrorDetail{
			Message: "Record not found",
			Type:    "NOT_FOUND",
		},
	}

	jsonData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if decoded.Error.Message != errorResp.Error.Message {
		t.Errorf("Error.Message = %q, want %q", decoded.Error.Message, errorResp.Error.Message)
	}

	if decoded.Error.Type != errorResp.Error.Type {
		t.Errorf("Error.Type = %q, want %q", decoded.Error.Type, errorResp.Error.Type)
	}
}

func TestErrorDetailJSONMarshaling(t *testing.T) {
	detail := ErrorDetail{
		Message: "Invalid request",
		Type:    "INVALID_REQUEST",
	}

	jsonData, err := json.Marshal(detail)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorDetail: %v", err)
	}

	var decoded ErrorDetail
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ErrorDetail: %v", err)
	}

	if decoded.Message != detail.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, detail.Message)
	}

	if decoded.Type != detail.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, detail.Type)
	}
}

func TestSimpleErrorResponseJSONMarshaling(t *testing.T) {
	errorResp := SimpleErrorResponse{
		Error: "Something went wrong",
	}

	jsonData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal SimpleErrorResponse: %v", err)
	}

	var decoded SimpleErrorResponse
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SimpleErrorResponse: %v", err)
	}

	if decoded.Error != errorResp.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, errorResp.Error)
	}
}

func TestAirtableRecordJSONFieldNames(t *testing.T) {
	jsonStr := `{"id":"rec123","createdTime":"2023-06-15T14:30:00Z","fields":{"name":"Test","value":42}}`

	var record AirtableRecord[TestFields]
	if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if record.Id != "rec123" {
		t.Errorf("Id = %q, want %q", record.Id, "rec123")
	}

	if record.Fields.Name != "Test" {
		t.Errorf("Fields.Name = %q, want %q", record.Fields.Name, "Test")
	}
}

func TestAirtableRecordsJSONFieldNames(t *testing.T) {
	jsonStr := `{"records":[{"id":"rec1","createdTime":"2023-01-01T00:00:00Z","fields":{"name":"First","value":1}}],"offset":"token123"}`

	var records AirtableRecords[TestFields]
	if err := json.Unmarshal([]byte(jsonStr), &records); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(records.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(records.Records))
	}

	if records.Offset != "token123" {
		t.Errorf("Offset = %q, want %q", records.Offset, "token123")
	}
}

func TestErrorResponseJSONFieldNames(t *testing.T) {
	jsonStr := `{"error":{"message":"Error message","type":"ERROR_TYPE"}}`

	var errorResp ErrorResponse
	if err := json.Unmarshal([]byte(jsonStr), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if errorResp.Error.Message != "Error message" {
		t.Errorf("Error.Message = %q, want %q", errorResp.Error.Message, "Error message")
	}

	if errorResp.Error.Type != "ERROR_TYPE" {
		t.Errorf("Error.Type = %q, want %q", errorResp.Error.Type, "ERROR_TYPE")
	}
}

func TestSimpleErrorResponseJSONFieldNames(t *testing.T) {
	jsonStr := `{"error":"Simple error"}`

	var errorResp SimpleErrorResponse
	if err := json.Unmarshal([]byte(jsonStr), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if errorResp.Error != "Simple error" {
		t.Errorf("Error = %q, want %q", errorResp.Error, "Simple error")
	}
}

func TestPatchItemsRequestJSONFieldNames(t *testing.T) {
	jsonStr := `{"records":[{"id":"rec1","fields":{"name":"Test","value":1}}],"typecast":true}`

	var request PatchItemsRequest[TestFields]
	if err := json.Unmarshal([]byte(jsonStr), &request); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if !request.Typecast {
		t.Error("Typecast should be true")
	}

	if len(request.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(request.Records))
	}
}

func TestCreateRecordsRequestJSONFieldNames(t *testing.T) {
	jsonStr := `{"records":[{"fields":{"name":"New","value":100}}],"typecast":false}`

	var request CreateRecordsRequest[TestFields]
	if err := json.Unmarshal([]byte(jsonStr), &request); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if request.Typecast {
		t.Error("Typecast should be false")
	}

	if len(request.Records) != 1 {
		t.Errorf("Records length = %d, want 1", len(request.Records))
	}
}

// Test empty structs
func TestEmptyStructs(t *testing.T) {
	// AirtableRecord
	record := AirtableRecord[TestFields]{}
	if record.Id != "" {
		t.Errorf("Empty AirtableRecord.Id should be empty string")
	}

	// AirtableRecords
	records := AirtableRecords[TestFields]{}
	if len(records.Records) != 0 {
		t.Errorf("Empty AirtableRecords.Records should have length 0")
	}
	if records.Offset != "" {
		t.Errorf("Empty AirtableRecords.Offset should be empty string")
	}

	// ErrorResponse
	errorResp := ErrorResponse{}
	if errorResp.Error.Message != "" {
		t.Errorf("Empty ErrorResponse.Error.Message should be empty string")
	}

	// SimpleErrorResponse
	simpleError := SimpleErrorResponse{}
	if simpleError.Error != "" {
		t.Errorf("Empty SimpleErrorResponse.Error should be empty string")
	}
}

// Benchmark tests
func BenchmarkAirtableRecordMarshal(b *testing.B) {
	record := AirtableRecord[TestFields]{
		Id:          "rec123abc",
		CreatedTime: time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC),
		Fields:      TestFields{Name: "Test", Value: 42},
	}

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(record)
	}
}

func BenchmarkAirtableRecordUnmarshal(b *testing.B) {
	jsonStr := `{"id":"rec123","createdTime":"2023-06-15T14:30:00Z","fields":{"name":"Test","value":42}}`
	data := []byte(jsonStr)

	for i := 0; i < b.N; i++ {
		var record AirtableRecord[TestFields]
		_ = json.Unmarshal(data, &record)
	}
}
