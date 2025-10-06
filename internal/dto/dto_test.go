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

func TestNormalizeName(t *testing.T) {
	t.Run("should normalize binary name", func(t *testing.T) {
		testCases := []struct {
			name       string
			binaryName string
			expected   string
		}{
			{
				name:       "full name (different owner and name)",
				binaryName: "foo/bar",
				expected:   "foo/bar",
			}, {
				name:       "same name and owner",
				binaryName: "foo",
				expected:   "foo/foo",
			}, {
				name:       "empty",
				binaryName: "",
				expected:   "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := NormalizeName(tc.binaryName)
				assert.Equal(t, tc.expected, got)
			})
		}
	})
}

func TestDisplayName(t *testing.T) {
	t.Run("should display binary name", func(t *testing.T) {
		testCases := []struct {
			name       string
			binaryInfo BinaryInfo
			expected   string
		}{
			{
				name: "full name (different owner and name)",
				binaryInfo: BinaryInfo{
					FullName:         "toto/tata",
					Name:             "tata",
					Owner:            "toto",
					InstalledVersion: "0.0.4",
				},
				expected: "toto/tata",
			}, {
				name: "same name and owner",
				binaryInfo: BinaryInfo{
					FullName:         "titi/titi",
					Name:             "titi",
					Owner:            "titi",
					InstalledVersion: "0.0.3",
				},
				expected: "titi",
			}, {
				name:       "empty",
				binaryInfo: BinaryInfo{},
				expected:   "",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got := tc.binaryInfo.DisplayName()
				assert.Equal(t, tc.expected, got)
			})
		}
	})
}
