package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
)

const (
	testBinaryName    = "foo"
	testBinaryVersion = "latest"
)

func TestNewState(t *testing.T) {
	t.Run("should init empty state", func(t *testing.T) {
		statePath := "foo.json"
		state := NewState(statePath)

		assert.NotNil(t, state)
		assert.Empty(t, state.Binaries)
		assert.Equal(t, statePath, state.path)
	})
}

func TestStateHelpers(t *testing.T) {
	state := NewState("foo.json")

	name, version := testBinaryName, testBinaryVersion
	binaryInfo := dto.BinaryInfo{FullName: name, Version: version, Name: name, Owner: name, InstalledVersion: version}

	require.Empty(t, state.Binaries)
	state.UpdateEntrie(binaryInfo)
	require.Len(t, state.Binaries, 1)
	assert.Equal(t, name, state.Binaries[name].Name)
	assert.Equal(t, name, state.Binaries[name].FullName)
	assert.Equal(t, name, state.Binaries[name].Owner)
	assert.Equal(t, version, state.Binaries[name].Version)
	assert.Equal(t, version, state.Binaries[name].InstalledVersion)

	ok := state.Has(name)
	assert.True(t, ok)
	ko := state.Has("notFound")
	assert.False(t, ko)
}

func TestLoad(t *testing.T) {
	t.Run("should create the file when not present and handle empty state", func(t *testing.T) {
		path := t.TempDir()
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)

		err := state.Load()
		assert.NoError(t, err)

		_, err = os.Stat(statePath)
		assert.NoError(t, err)
		assert.Empty(t, state.Binaries)
	})

	t.Run("should lock the state file", func(t *testing.T) {
		path := t.TempDir()
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)

		err := state.Load()
		assert.NoError(t, err)

		err = state.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "try again later")
	})

	t.Run("should load the state file", func(t *testing.T) {
		path := t.TempDir()
		name, version := testBinaryName, testBinaryVersion
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)

		data := make([]dto.BinaryInfo, 1)
		data[0] = dto.BinaryInfo{FullName: name, Version: version, Name: name, Owner: name, InstalledVersion: version}
		file, err := os.Create(filepath.Clean(statePath))
		require.NoError(t, err)
		err = json.NewEncoder(file).Encode(data)
		require.NoError(t, err)

		err = state.Load()
		require.NoError(t, err)
		require.Len(t, state.Binaries, 1)
		assert.Equal(t, name, state.Binaries[name].Name)
		assert.Equal(t, name, state.Binaries[name].FullName)
		assert.Equal(t, name, state.Binaries[name].Owner)
		assert.Equal(t, version, state.Binaries[name].Version)
		assert.Equal(t, version, state.Binaries[name].InstalledVersion)
	})

	t.Run("should return error on bad json", func(t *testing.T) {
		path := t.TempDir()
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)

		err := os.WriteFile(statePath, []byte("{foo:"), 0o600)
		require.NoError(t, err)

		err = state.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "object key string")
	})

	t.Run("should return error on invalid open", func(t *testing.T) {
		path := t.TempDir()
		invalidPath := filepath.Join(path, "nonexistent", "state.json")
		state := NewState(invalidPath)

		err := state.Load()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestSave(t *testing.T) {
	t.Run("should save cache", func(t *testing.T) {
		path := t.TempDir()
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)
		name, version := testBinaryName, testBinaryVersion
		binaryInfo := dto.BinaryInfo{FullName: name, Version: version, Name: name, Owner: name, InstalledVersion: version}

		state.UpdateEntrie(binaryInfo)
		err := state.Save()
		require.NoError(t, err)

		input, err := os.ReadFile(statePath)
		require.NoError(t, err)
		stateData := make([]dto.BinaryInfo, 1)
		err = json.Unmarshal(input, &stateData)
		require.NoError(t, err)
		require.Len(t, stateData, 1)
		assert.Equal(t, binaryInfo, stateData[0])
	})

	t.Run("should unlock the file after save", func(t *testing.T) {
		path := t.TempDir()
		statePath := filepath.Join(path, "state.json")
		state := NewState(statePath)
		name, version := testBinaryName, testBinaryVersion
		binaryInfo := dto.BinaryInfo{FullName: name, Version: version, Name: name, Owner: name, InstalledVersion: version}

		err := state.Load()
		require.NoError(t, err)

		state.UpdateEntrie(binaryInfo)
		err = state.Save()
		require.NoError(t, err)

		// file should be unlocked after save
		err = state.Load()
		require.NoError(t, err)
	})

	t.Run("should return an error on invalid path", func(t *testing.T) {
		path := t.TempDir()
		invalidPath := filepath.Join(path, "nonexistent", "state.json")
		state := NewState(invalidPath)

		err := state.Save()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})
}

func TestStateDirectory(t *testing.T) {
	t.Run("should return user config dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
		t.Setenv("HOME", tmpDir)

		userConfigDir, err := os.UserConfigDir()
		require.NoError(t, err)

		got := StateDirectory()
		assert.Equal(t, userConfigDir+"/azabox", got)
	})
}
