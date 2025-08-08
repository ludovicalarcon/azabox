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

func TestSetFromSlice(t *testing.T) {
	t.Run("should create a new set from a slice", func(t *testing.T) {
		sliceString := []string{"foo", "bar"}
		sliceInt := []int{10, 42, 7, 19}

		setS := SetFromSlice(sliceString)
		setI := SetFromSlice(sliceInt)

		require.NotNil(t, setS)
		require.NotNil(t, setI)
		assert.Len(t, setS, len(sliceString))
		assert.Len(t, setI, len(sliceInt))
		for key := range setS {
			assert.Contains(t, sliceString, key, fmt.Sprintf("item %s should be in Set ", key))
		}
		for key := range setI {
			assert.Contains(t, sliceInt, key, fmt.Sprintf("item %d should be in Set ", key))
		}
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

		set.Add(expected[0], expected[1], expected[2], expected[3])

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

func TestSet_Remove(t *testing.T) {
	t.Run("should remove single item", func(t *testing.T) {
		set := NewSet[int](4)
		set.Add(42, 10, 1, 19)

		set.Remove(10)
		assert.Len(t, set, 3)
		assert.False(t, set.Has(10))
	})

	t.Run("should remove single item", func(t *testing.T) {
		set := NewSet[int](4)
		set.Add(42, 10, 1, 19)

		set.Remove(10, 19)
		assert.Len(t, set, 2)
		assert.False(t, set.Has(10))
		assert.False(t, set.Has(19))
	})

	t.Run("should handle empty", func(t *testing.T) {
		set := NewSet[int]()
		set.Remove(10)
		assert.Len(t, set, 0)
	})
}

func TestSet_Len(t *testing.T) {
	t.Run("should return set len", func(t *testing.T) {
		testCases := []struct {
			name string
			set  Set[int]

			expected int
		}{
			{
				name:     "empty",
				set:      NewSet[int](),
				expected: 0,
			},
			{
				name:     "single item",
				set:      SetFromSlice([]int{42}),
				expected: 1,
			},
			{
				name:     "multiple item",
				set:      SetFromSlice([]int{42, 1, 19, 7, 14}),
				expected: 5,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.set.Len())
			})
		}
	})
}

func TestSet_Empty(t *testing.T) {
	t.Run("should return true on empty set", func(t *testing.T) {
		set := NewSet[int]()
		assert.True(t, set.Empty())
	})

	t.Run("should return false on non empty set", func(t *testing.T) {
		set := NewSet[int]()
		set.Add(42, 10)
		assert.False(t, set.Empty())
	})
}

func TestSet_Clear(t *testing.T) {
	t.Run("should empty a set", func(t *testing.T) {
		set := NewSet[string](3)
		set.Add("foo", "bar", "toto")

		set.Clear()

		assert.Empty(t, set)
	})

	t.Run("should handle empty set", func(t *testing.T) {
		set := NewSet[string]()

		set.Clear()

		assert.Empty(t, set)
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

func TestSet_String(t *testing.T) {
	t.Run("should stringify set", func(t *testing.T) {
		testCases := []struct {
			name     string
			set      Set[string]
			expected string
		}{
			{
				name:     "empty",
				set:      NewSet[string](),
				expected: "{}",
			},
			{
				name:     "nil",
				set:      nil,
				expected: "{}",
			},
			// using single values as underling datastructure is a map
			// and in go map doesn't have a deterministic order
			{
				name:     "values",
				set:      SetFromSlice([]string{"foo"}),
				expected: "{foo bar toto}",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.Equal(t, tc.expected, tc.set.String())
			})
		}
	})
}
