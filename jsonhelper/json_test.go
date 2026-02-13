package jsonhelper

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type rawStruct struct {
	Raw string `json:"raw"`
}
type inputStruct struct {
	Content rawStruct `json:"content"`
}

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
				if err == nil || tt.expectedError == nil {
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

func TestWriteToJSONFile_SuccessfulWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	input := inputStruct{
		Content: rawStruct{
			Raw: "test content",
		},
	}

	err := WriteToJSONFile(filePath, input, false)
	if err != nil {
		t.Fatalf("WriteToJSONFile() returned unexpected error: %v", err)
	}

	// Verify file was created and contains expected content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := `{"content":{"raw":"test content"}}`
	actual := strings.TrimSpace(string(data))
	if actual != expected {
		t.Errorf("Expected [%s] but got [%s]", expected, actual)
	}
}

func TestWriteToJSONFile_Overwrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.json")

	// Write initial content
	initialInput := inputStruct{
		Content: rawStruct{
			Raw: "initial",
		},
	}
	err := WriteToJSONFile(filePath, initialInput, false)
	if err != nil {
		t.Fatalf("Initial WriteToJSONFile() returned unexpected error: %v", err)
	}

	// Overwrite with new content
	newInput := inputStruct{
		Content: rawStruct{
			Raw: "overwritten",
		},
	}
	err = WriteToJSONFile(filePath, newInput, true)
	if err != nil {
		t.Fatalf("Overwrite WriteToJSONFile() returned unexpected error: %v", err)
	}

	// Verify file contains new content
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := `{"content":{"raw":"overwritten"}}`
	actual := strings.TrimSpace(string(data))
	if actual != expected {
		t.Errorf("Expected [%s] but got [%s]", expected, actual)
	}
}

func TestGetJSONString(t *testing.T) {
	var inputStruct = inputStruct{
		Content: rawStruct{
			Raw: "something",
		},
	}
	var array = []string{"a", "b", "c"}

	var tests = []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"empty", nil, "null"},
		{"simple_object", inputStruct, "{\"content\":{\"raw\":\"something\"}}"},
		{"simple_array", array, "[\"a\",\"b\",\"c\"]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := GetJSONString(tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			trimmed := strings.ReplaceAll(actual, "\n", "")
			if trimmed != tt.expected {
				t.Errorf("Expected [%s] but got [%s]", tt.expected, trimmed)
			}
		})
	}
}

func TestParseJSONReader(t *testing.T) {
	var inputData = inputStruct{
		Content: rawStruct{
			Raw: "something",
		},
	}

	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"simple_object", "{\"content\":{\"raw\":\"something\"}}", inputData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual inputStruct
			buf := bytes.NewBufferString(tt.input)
			err := ParseJSONReader(buf, &actual)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := tt.expected.(inputStruct)
			if actual.Content.Raw != expected.Content.Raw {
				t.Errorf("Expected [%v] but got [%v]", tt.expected, actual)
			}
		})
	}

}

func TestParseJSONString(t *testing.T) {
	var inputData = inputStruct{
		Content: rawStruct{
			Raw: "something",
		},
	}

	var tests = []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"simple_object", "{\"content\":{\"raw\":\"something\"}}", inputData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual inputStruct
			err := ParseJSONString(tt.input, &actual)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			expected := tt.expected.(inputStruct)
			if actual.Content.Raw != expected.Content.Raw {
				t.Errorf("Expected [%v] but got [%v]", tt.expected, actual)
			}
		})
	}

}

func TestGetJSONStringBuffer(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil_value", nil, "null"},
		{"simple_string", "hello", `"hello"`},
		{"number", 42, "42"},
		{"boolean", true, "true"},
		{"struct", rawStruct{Raw: "test"}, `{"raw":"test"}`},
		{"map", map[string]int{"a": 1, "b": 2}, `{"a":1,"b":2}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := GetJSONStringBuffer(tt.input)
			if err != nil {
				t.Fatalf("GetJSONStringBuffer() returned unexpected error: %v", err)
			}
			if buf == nil {
				t.Fatal("GetJSONStringBuffer() returned nil buffer")
			}
			actual := strings.TrimSpace(buf.String())
			if actual != tt.expected {
				t.Errorf("Expected [%s] but got [%s]", tt.expected, actual)
			}
		})
	}
}

func TestGetJSONStringBuffer_Error(t *testing.T) {
	// Channels cannot be marshaled to JSON
	ch := make(chan int)
	_, err := GetJSONStringBuffer(ch)
	if err == nil {
		t.Error("Expected error for unmarshallable type, got nil")
	}
}

func TestParseJSONReader_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid_json", "{invalid json}"},
		{"unclosed_brace", `{"key": "value"`},
		{"empty_string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output inputStruct
			buf := bytes.NewBufferString(tt.input)
			err := ParseJSONReader(buf, &output)
			if err == nil {
				t.Errorf("Expected error for invalid JSON [%s], got nil", tt.input)
			}
		})
	}
}

func TestParseJSONString_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid_json", "{invalid json}"},
		{"unclosed_brace", `{"key": "value"`},
		{"wrong_type", `["array", "not", "object"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output inputStruct
			err := ParseJSONString(tt.input, &output)
			if err == nil {
				t.Errorf("Expected error for invalid JSON [%s], got nil", tt.input)
			}
		})
	}
}

func TestParseJSONFromBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected inputStruct
	}{
		{
			"simple_object",
			[]byte(`{"content":{"raw":"something"}}`),
			inputStruct{Content: rawStruct{Raw: "something"}},
		},
		{
			"empty_content",
			[]byte(`{"content":{"raw":""}}`),
			inputStruct{Content: rawStruct{Raw: ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual inputStruct
			err := ParseJSONFromBytes(&actual, tt.input)
			if err != nil {
				t.Fatalf("ParseJSONFromBytes() returned unexpected error: %v", err)
			}
			if actual.Content.Raw != tt.expected.Content.Raw {
				t.Errorf("Expected [%v] but got [%v]", tt.expected, actual)
			}
		})
	}
}

func TestParseJSONFromBytes_Error(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"invalid_json", []byte("{invalid json}")},
		{"unclosed_brace", []byte(`{"key": "value"`)},
		{"empty_bytes", []byte("")},
		{"nil_bytes", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output inputStruct
			err := ParseJSONFromBytes(&output, tt.input)
			if err == nil {
				t.Errorf("Expected error for invalid JSON [%s], got nil", string(tt.input))
			}
		})
	}
}
