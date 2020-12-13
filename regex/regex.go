package regex

import "regexp"

// FindNamedGroupMatchedStrings returns a map with named groups specified in the
// specified regex
func FindNamedGroupMatchedStrings(regex *regexp.Regexp, input string) map[string]string {
	match := regex.FindStringSubmatch(input)
	subMatchMap := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 {
			subMatchMap[name] = match[i]
		}
	}

	return subMatchMap
}
