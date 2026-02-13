package datetime

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGetLocalDateTimeString(t *testing.T) {
	// Use local timezone for test inputs to avoid timezone conversion issues
	loc, _ := time.LoadLocation("Local")

	tests := []struct {
		name  string
		input time.Time
	}{
		{
			name:  "local time",
			input: time.Date(2024, 6, 15, 14, 30, 0, 0, loc),
		},
		{
			name:  "midnight local",
			input: time.Date(2024, 1, 1, 0, 0, 0, 0, loc),
		},
		{
			name:  "end of day local",
			input: time.Date(2024, 3, 31, 23, 59, 59, 0, loc),
		},
		{
			name:  "UTC time",
			input: time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLocalDateTimeString(&tt.input)

			// Check format is "YYYY-MM-DD HH:MM"
			if len(result) != 16 {
				t.Errorf("expected result length 16 (YYYY-MM-DD HH:MM), got %d: %q", len(result), result)
			}

			// Verify the format pattern
			if result[4] != '-' || result[7] != '-' || result[10] != ' ' || result[13] != ':' {
				t.Errorf("result does not match expected format YYYY-MM-DD HH:MM: %q", result)
			}

			// Parse the result back to verify it's a valid datetime string
			_, err := time.Parse("2006-01-02 15:04", result)
			if err != nil {
				t.Errorf("result %q is not a valid datetime string: %v", result, err)
			}
		})
	}
}

func TestGetLocalDateTimeString_VerifyConversion(t *testing.T) {
	// Test that the function correctly converts to local time
	loc, _ := time.LoadLocation("Local")
	localTime := time.Date(2024, 6, 15, 14, 30, 0, 0, loc)

	result := GetLocalDateTimeString(&localTime)
	expected := "2024-06-15 14:30"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		expectedError error
	}{
		{
			name:          "valid RFC3339 date",
			str:           "2022-11-21T05:00:00Z",
			expectedError: nil,
		},
		{
			name:          "invalid RFC3339 date",
			str:           "2022-11-21",
			expectedError: errors.New("invalid date format"),
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf(test.name, test.str)
		t.Run(testName, func(t *testing.T) {
			actualErr := ValidateDate(test.str)
			if test.expectedError == nil {
				if actualErr != nil {
					t.Errorf("expected no error, got %v", actualErr)
				}
				return
			}
			if actualErr == nil {
				t.Errorf("expected error containing %q, got nil", test.expectedError.Error())
				return
			}
			if !strings.Contains(actualErr.Error(), test.expectedError.Error()) {
				t.Errorf("expected error containing %q, got %v", test.expectedError.Error(), actualErr)
			}
		})
	}
}

func TestValidateDate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid RFC3339 with timezone offset",
			input:       "2024-06-15T14:30:00+05:30",
			expectError: false,
		},
		{
			name:        "valid RFC3339 with negative offset",
			input:       "2024-06-15T14:30:00-08:00",
			expectError: false,
		},
		{
			name:        "valid RFC3339 with Z timezone",
			input:       "2024-01-01T00:00:00Z",
			expectError: false,
		},
		{
			name:        "valid RFC3339 with milliseconds",
			input:       "2024-06-15T14:30:00.123Z",
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "date only without time",
			input:       "2024-06-15",
			expectError: true,
		},
		{
			name:        "time only without date",
			input:       "14:30:00",
			expectError: true,
		},
		{
			name:        "invalid month",
			input:       "2024-13-15T14:30:00Z",
			expectError: true,
		},
		{
			name:        "invalid day",
			input:       "2024-06-32T14:30:00Z",
			expectError: true,
		},
		{
			name:        "invalid format - missing T separator",
			input:       "2024-06-15 14:30:00Z",
			expectError: true,
		},
		{
			name:        "random string",
			input:       "not a date",
			expectError: true,
		},
		{
			name:        "Unix timestamp",
			input:       "1718451000",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error for input %q, got %v", tt.input, err)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid date format") {
				t.Errorf("expected error message to contain 'invalid date format', got %v", err)
			}
		})
	}
}
