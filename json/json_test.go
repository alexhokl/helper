package json

import (
	"errors"
	"strings"
	"testing"
)

func TestWriteToJSONFile(t *testing.T) {
	var tests = []struct {
		name          string
		path          string
		input         interface{}
		isOverwrite   bool
		expected      string
		expectedError error
	}{
		{"empty_path", "", []string{}, false, "", errors.New("path is not specified")},
		{"empty_object", "/home/user/obj.json", nil, false, "", errors.New("object cannot be empty")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteToJSONFile(tt.path, tt.input, tt.isOverwrite)
			if err != nil || tt.expectedError != nil {
				if !(err != nil && tt.expectedError != nil) {
					t.Errorf("Expected error [%v] but got [%v]", tt.expectedError, err)
					return
				}
				if err.Error() != tt.expectedError.Error() {
					t.Errorf("Expected error [%s] but got [%s]", tt.expectedError.Error(), err.Error())
				}
			}
			// if len(actual) != len(tt.expected) {
			// 	t.Errorf("Expected %v but got %v", tt.expected, actual)
			// }
		})
	}

}

func TestGetJSONString(t *testing.T) {
	type RawStruct struct {
		Raw string `json:"raw"`
	}
	var inputStruct = struct {
		Content RawStruct `json:"content"`
	}{
		Content: RawStruct{
			Raw: "something",
		},
	}

	var tests = []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"empty", nil, "null"},
		{"simple_object", inputStruct, "{\"content\":{\"raw\":\"something\"}}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetJSONString(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %w", err)
			}
			trimmed := strings.ReplaceAll(actual, "\n", "")
			if trimmed != tt.expected {
				t.Errorf("Expected [%s] but got [%s]", tt.expected, trimmed)
			}
		})
	}
}
