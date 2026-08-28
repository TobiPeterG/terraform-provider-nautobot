package shared

func SliceToSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			set[value] = struct{}{}
		}
	}
	return set
}

// SetDiff returns items in a that are not in b.
func SetDiff(a, b map[string]struct{}) []string {
	difference := make([]string, 0, len(a))
	for value := range a {
		if _, exists := b[value]; !exists {
			difference = append(difference, value)
		}
	}
	return difference
}
