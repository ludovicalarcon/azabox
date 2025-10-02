package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

func createFakeState(binaries []dto.BinaryInfo) *DummyState {
	dummyState := &DummyState{
		binaries: make(map[string]dto.BinaryInfo, len(binaries)),
	}

	for _, binaryInfo := range binaries {
		dummyState.UpdateEntrie(binaryInfo)
	}
	return dummyState
}

func TestNewListCommand(t *testing.T) {
	t.Run("should create a new list command", func(t *testing.T) {
		dummyState := &DummyState{}
		cmd := newListCommand(dummyState)

		require.NotNil(t, cmd)
		assert.Equal(t, ListUseMessage, cmd.Use)
		assert.Equal(t, ListShortMessage, cmd.Short)
		assert.NotNil(t, cmd.RunE)
		assert.True(t, cmd.SilenceErrors)
		assert.True(t, cmd.SilenceUsage)
	})
}

func TestExecuteListCommand(t *testing.T) {
	t.Run("should list binaries from state", func(t *testing.T) {
		testCases := []struct {
			name     string
			binaries []dto.BinaryInfo
			expected []string
		}{
			{
				name: "single binary",
				binaries: []dto.BinaryInfo{
					{FullName: "foo/foo", Name: "foo", Owner: "foo", InstalledVersion: "0.0.1"},
				},
				expected: []string{"foo in version 0.0.1"},
			},
			{
				name: "multiple binaries",
				binaries: []dto.BinaryInfo{
					{FullName: "foo/foo", Name: "foo", Owner: "foo", InstalledVersion: "0.0.1"},
					{FullName: "foo/bar", Name: "bar", Owner: "foo", InstalledVersion: "0.0.2"},
				},
				expected: []string{
					"foo in version 0.0.1",
					"foo/bar in version 0.0.2",
				},
			},
			{
				name:     "empty state",
				binaries: []dto.BinaryInfo{},
				expected: []string{"No binary installed"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				dummyState := createFakeState(tc.binaries)

				got, errList := executeListCommand(dummyState)
				require.NoError(t, errList)

				for _, expected := range tc.expected {
					assert.Contains(t, got, expected)
				}
			})
		}
	})

	t.Run("should handle error", func(t *testing.T) {
		dummyState := createFakeState([]dto.BinaryInfo{})
		dummyState.onError = true

		got, errCmd := executeListCommand(dummyState)
		require.Error(t, errCmd)
		assert.Empty(t, got)

		cmd := newListCommand(dummyState)
		errCmd = cmd.RunE(cmd, []string{})
		assert.Error(t, errCmd)
	})
}
