package types

import (
	"fmt"
	"strings"
)

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](size ...int) Set[T] {
	if len(size) > 0 {
		return make(Set[T], size[0])
	}
	return make(Set[T])
}

func SetFromSlice[T comparable](items []T) Set[T] {
	set := NewSet[T](len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func (s Set[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set[T]) Has(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) Len() int { return len(s) }

func (s Set[T]) Empty() bool { return s.Len() == 0 }

func (s Set[T]) Clear() {
	for item := range s {
		delete(s, item)
	}
}

func (s Set[T]) ToSlice() []T {
	slice := make([]T, 0, len(s))
	for key := range s {
		slice = append(slice, key)
	}
	return slice
}

func (s Set[T]) String() string {
	if s == nil || s.Empty() {
		return "{}"
	}

	var sb strings.Builder
	i, length := 0, s.Len()
	sb.WriteString("{")
	for item := range s {
		sb.WriteString(fmt.Sprintf("%v", item))
		if i < length-1 {
			sb.WriteString(" ")
		}
		i++
	}
	sb.WriteString("}")
	return sb.String()
}
