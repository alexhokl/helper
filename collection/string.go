package collection

import (
	"fmt"
	"strings"
)

// GetDistinct returns a distinct (non-duplicated) array from the specified input
func GetDistinct(array []string) []string {
	m := map[string]bool{}

	for v := range array {
		m[array[v]] = true
	}

	d := []string{}
	for k := range m {
		d = append(d, k)
	}
	return d
}

// GetDelimitedString returns a delimited string using a specified array
func GetDelimitedString(array []string, delimiter string) string {
	var builder strings.Builder
	for index, i := range array {
		if index == 0 {
			builder.WriteString(fmt.Sprintf("%s", i))
		} else {
			builder.WriteString(fmt.Sprintf("%s%s", delimiter, i))
		}
	}
	return builder.String()
}
