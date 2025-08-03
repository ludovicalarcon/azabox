package types

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](size ...int) Set[T] {
	if len(size) > 0 {
		return make(Set[T], size[0])
	}
	return make(Set[T])
}

func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s))
	for key := range s {
		slice = append(slice, key)
	}
	return slice
}
