package leif

// Union of two arrays of strings, used to help compute the merge rules
func union(a, b []string) []string {
	m := make(map[string]bool)
	for _, str := range a {
		m[str] = true
	}

	for _, str := range b {
		if _, ok := m[str]; !ok {
			a = append(a, str)
		}
	}
	return a
}

// Difference between a and b, so the set equivalent of a - b
func difference(a, b []string) []string {
	var diff []string
	m := make(map[string]bool)
	for _, str := range b {
		m[str] = true
	}

	for _, str := range a {
		if _, ok := m[str]; !ok {
			diff = append(diff, str)
		}
	}
	return diff
}
