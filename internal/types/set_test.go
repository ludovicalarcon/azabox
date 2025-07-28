package types

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSet(t *testing.T) {
	t.Run("should init empty set", func(t *testing.T) {
		set := NewSet[int]()
		require.NotNil(t, set)
		assert.Len(t, set, 0)
	})
}

func TestSet_Add(t *testing.T) {
	t.Run("should add item", func(t *testing.T) {
		set := NewSet[int]()
		expected := 10

		set.Add(expected)

		assert.Len(t, set, 1)
		_, ok := set[expected]
		assert.True(t, ok, fmt.Sprintf("item %d should be in Set", expected))
	})

	t.Run("should add multiple items", func(t *testing.T) {
		set := NewSet[string]()
		expected := []string{"toto", "tata", "titi", "tutu"}

		for _, item := range expected {
			set.Add(item)
		}

		assert.Len(t, set, len(expected))
		for key := range set {
			assert.Contains(t, expected, key, fmt.Sprintf("item %s should be in Set ", key))
		}
	})

	t.Run("should handle duplicate", func(t *testing.T) {
		set := NewSet[string]()
		duplicate := "foo"
		expected := []string{"bar", duplicate}

		set.Add(duplicate)
		set.Add(expected[0])
		set.Add(duplicate)

		assert.Len(t, set, len(expected))
		for key := range set {
			assert.Contains(t, expected, key, fmt.Sprintf("item %s should be in Set ", key))
		}
	})
}

func TestSet_Has(t *testing.T) {
	t.Run("should tell if item exist", func(t *testing.T) {
		testCases := []struct {
			name        string
			set         Set[string]
			items       []string
			assertItems []string
			expected    []bool
		}{
			{
				name:        "single item",
				set:         NewSet[string](1),
				items:       []string{"foo"},
				assertItems: []string{"foo"},
				expected:    []bool{true},
			},
			{
				name:        "multiple item",
				set:         NewSet[string](3),
				items:       []string{"foo", "bar", "bob"},
				assertItems: []string{"foo", "bar", "unkn", "bob"},
				expected:    []bool{true, true, false, true},
			},
			{
				name:        "empty",
				set:         NewSet[string](0),
				items:       []string{},
				assertItems: []string{},
				expected:    []bool{false},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				for _, item := range tc.items {
					tc.set.Add(item)
				}
				for i, assertItem := range tc.assertItems {
					assert.Equal(t, tc.expected[i], tc.set.Has(assertItem))
				}
			})
		}
	})
}

func TestSet_ToSlice(t *testing.T) {
	t.Run("should convert to slice", func(t *testing.T) {
		testCases := []struct {
			name     string
			items    []int
			expected []int
		}{
			{
				name:     "empty set",
				items:    []int{},
				expected: []int{},
			},
			{
				name:     "single item",
				items:    []int{42},
				expected: []int{42},
			},
			{
				name:     "multiple items",
				items:    []int{42, 11, 7, 19},
				expected: []int{42, 11, 7, 19},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				set := NewSet[int](len(tc.items))

				for _, item := range tc.items {
					set.Add(item)
				}

				slice := set.ToSlice()
				sort.Ints(slice)
				sort.Ints(tc.expected)

				require.Len(t, slice, len(tc.expected))
				assert.Equal(t, tc.expected, slice)
			})
		}
	})
}
