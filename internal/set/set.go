package set

import (
	"iter"
)

type Set[E comparable] struct {
	m map[E]struct{}
}

func Make[E comparable]() *Set[E] {
	return &Set[E]{
		m: make(map[E]struct{}),
	}
}

func (s *Set[E]) Add(v E) {
	s.m[v] = struct{}{}
}

func (s *Set[E]) Delete(v E) {
	delete(s.m, v)
}

func (s *Set[E]) Contains(v E) bool {
	_, ok := s.m[v]
	return ok
}

func (s *Set[E]) Len() int {
	return len(s.m)
}

func (s *Set[E]) All() iter.Seq[E] {
	return func(yield func(E) bool) {
		for v := range s.m {
			if !yield(v) {
				break
			}
		}
	}
}
