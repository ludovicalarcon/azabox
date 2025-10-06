package platform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeArch(t *testing.T) {
	t.Run("should normalize architecture", func(t *testing.T) {
		testCases := []struct {
			name     string
			arch     string
			expected string
		}{
			{
				name:     "amd64",
				arch:     "amd64",
				expected: "x86_64",
			},
			{
				name:     "arm64",
				arch:     "arm64",
				expected: "aarch64",
			},
			{
				name:     "default",
				arch:     "darwin",
				expected: "darwin",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := NormalizeArch(tc.arch)
				assert.Equal(t, tc.expected, got)
			})
		}
	})
}
