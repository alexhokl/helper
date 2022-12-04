package datetime

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

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
			if actualErr != test.expectedError {
				if !strings.Contains(actualErr.Error(), test.expectedError.Error()) {
					t.Errorf("expected error %v, got %v", test.expectedError, actualErr)
				}
				return
			}
		})
	}
}

