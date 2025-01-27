package collection

import "testing"

func TestGetDistinct(t *testing.T) {
	var tests = []struct {
		name            string
		input, expected []string
	}{
		{"empty", []string{}, []string{}},
		{"non-duplicated", []string{"1", "2", "3", "5", "4"}, []string{"1", "2", "3", "5", "4"}},
		{"duplicated", []string{"1", "2", "4", "3", "5", "3", "1", "4"}, []string{"1", "2", "4", "3", "5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetDistinct(tt.input)
			if len(actual) != len(tt.expected) {
				t.Errorf("Expected %v but got %v", tt.expected, actual)
			}
		})
	}
}

func TestGetDelimitedString(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		delimiter string
		expected  string
	}{
		{
			name:      "empty array",
			input:     []string{},
			delimiter: ",",
			expected:  "",
		},
		{
			name:      "empty delimiter",
			input:     []string{"abc", "def"},
			delimiter: "",
			expected:  "abcdef",
		},
		{
			name:      "normal case",
			input:     []string{"abc", "def"},
			delimiter: ",",
			expected:  "abc,def",
		},
		{
			name:      "single item array",
			input:     []string{"abc"},
			delimiter: ",",
			expected:  "abc",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := GetDelimitedString(test.input, test.delimiter)
			if actual != test.expected {
				t.Errorf("Expected %v but got %v", test.expected, actual)
			}
		})
	}
}

func BenchmarkGetDelimitedString(t *testing.B) {
	for i := 0; i < t.N; i++ {
		GetDelimitedString([]string{"abc", "def"}, ",")
	}
}
