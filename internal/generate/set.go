package generate

import (
	"github.com/lftk/anki-vocab/internal/set"
)

type linkedSet[K comparable, V any] struct {
	seen *set.Set[K]
	vals []V
}

func (s *linkedSet[K, V]) addFunc(key K, valf func() (V, error)) (bool, error) {
	if s.seen == nil {
		s.seen = set.Make[K]()
	}

	if s.seen.Contains(key) {
		return false, nil
	}

	val, err := valf()
	if err != nil {
		return false, err
	}

	s.seen.Add(key)
	s.vals = append(s.vals, val)

	return true, nil
}

func (s *linkedSet[K, V]) values() []V {
	return s.vals
}
