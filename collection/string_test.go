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
