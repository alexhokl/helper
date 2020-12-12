package collection

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
