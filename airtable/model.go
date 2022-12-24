package airtable

import (
	"bytes"
	"time"
)

// AirtableFields a marker interface
// where any struct would work
type AirtableFields interface {
}

// AirtableRecords represents a list of records of Airtable
type AirtableRecords[T AirtableFields] struct {
	Records []AirtableRecord[T] `json:"records"`
	Offset  string              `json:"offset"`
}

// AirtableRecord represents a record of Airtable
type AirtableRecord[T AirtableFields] struct {
	Id          string    `json:"id"`
	CreatedTime time.Time `json:"createdTime"`
	Fields      T         `json:"fields"`
}

// PatchItemsRequest represents a request to patch a list of items
type PatchItemsRequest[T AirtableFields] struct {
	Records []PatchItemRequest[T] `json:"records"`
}

// PatchItemRequest represents a request to patch an item
type PatchItemRequest[T AirtableFields] struct {
	Id     string `json:"id"`
	Fields T      `json:"fields"`
}

// CreateRecordsRequest represents a request to create a list of records
type CreateRecordsRequest[T AirtableFields] struct {
	Records []CreateRecordRequest[T] `json:"records"`
}

// CreateRecordRequest represents a request to create a record
type CreateRecordRequest[T AirtableFields] struct {
	Fields T `json:"fields"`
}

// ErrorResponse represents an error response from Airtable
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents an error detail from Airtable
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// SimpleErrorResponse represents a simple error response from Airtable
type SimpleErrorResponse struct {
	Error string `json:"error"`
}

// PatchBooleanBody returns bytes of a request body to patch a boolean field
type PatchBooleanBody func(map[string]bool) (*bytes.Buffer, error)

// PatchStringBody returns bytes of a request body to patch a string field
type PatchStringBody func(map[string]string) (*bytes.Buffer, error)

// PatchStringArrayBody returns bytes of a request body to patch a string array field
type PatchStringArrayBody func(map[string][]string) (*bytes.Buffer, error)

// PatchTimeBody returns bytes of a request body to patch a time field
type PatchTimeBody func(map[string]time.Time) (*bytes.Buffer, error)
