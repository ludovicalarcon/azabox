package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinaryInfoString(t *testing.T) {
	t.Run("should format binary info", func(t *testing.T) {
		testCases := []struct {
			name       string
			binaryInfo BinaryInfo
			expected   string
		}{
			{
				name: "same name and owner",
				binaryInfo: BinaryInfo{
					FullName:         "foo/foo",
					Name:             "foo",
					Owner:            "foo",
					InstalledVersion: "0.0.1",
				},
				expected: "foo in version 0.0.1",
			}, {
				name: "different name and owner",
				binaryInfo: BinaryInfo{
					FullName:         "foo/bar",
					Name:             "bar",
					Owner:            "foo",
					InstalledVersion: "0.0.2",
				},
				expected: "foo/bar in version 0.0.2",
			}, {
				name:       "empty",
				binaryInfo: BinaryInfo{},
				expected:   "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := tc.binaryInfo.String()
				assert.Equal(t, tc.expected, got)
			})
		}
	})
}
