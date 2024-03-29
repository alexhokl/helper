package regexhelper

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFindNamedGroupMatchedStrings(t *testing.T) {
	tests := []struct {
		regex    string
		input    string
		expected map[string]string
	}{
		{
			regex: `bitbucket\.org/(?P<org>\w+)\/(?P<name>.*)`,
			input: "bitbucket.org/username/reponame",
			expected: map[string]string{
				"org":  "username",
				"name": "reponame",
			},
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("regex=%s, input=%s", test.regex, test.input)
		t.Run(testName, func(t *testing.T) {
			actual := FindNamedGroupMatchedStrings(regexp.MustCompile(test.regex), test.input)
			for k, v := range actual {
				if test.expected[k] != v {
					t.Errorf("expected key [%s] with value %s, got %v", k, test.expected[k], v)
				}
			}
		})
	}
}
