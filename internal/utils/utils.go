package utils

func SliceUnique[Slice ~[]E, E comparable](s Slice) Slice {
	seen := make(map[E]struct{}, len(s))
	s2 := make(Slice, 0, len(s))
	for _, e := range s {
		if _, ok := seen[e]; !ok {
			seen[e] = struct{}{}
			s2 = append(s2, e)
		}
	}
	return s2
}
