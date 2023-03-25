package jsonhelper

import (
	"bytes"
	"errors"
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
				if (err == nil || tt.expectedError == nil) {
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
