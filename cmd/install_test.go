package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/installer"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
)

func initLoggerForTest() {
	_ = logging.InitLogger(logging.Config{Encoding: logging.Json})
}

func TestNewInstallCommand(t *testing.T) {
	t.Run("should create a new install command", func(t *testing.T) {
		localInstaller, err := installer.New()
		dummyState := &DummyState{}
		require.NoError(t, err)
		require.NotNil(t, localInstaller)

		cmd := newInstallCommand(localInstaller, dummyState)

		require.NotNil(t, cmd)
		assert.Equal(t, InstallUseMessage, cmd.Use)
		assert.Equal(t, InstallShortMessage, cmd.Short)
		assert.NotNil(t, cmd.RunE)
		assert.True(t, cmd.SilenceErrors)
	})
}

func TestInstallCommand(t *testing.T) {
	t.Run("should return an error when there is less than one args provided", func(t *testing.T) {
		localInstaller, err := installer.New()
		dummyState := &DummyState{}
		require.NoError(t, err)
		require.NotNil(t, localInstaller)

		cmd := newInstallCommand(localInstaller, dummyState)
		err = cmd.RunE(cmd, []string{})
		require.Error(t, err)
		assert.Equal(t, ArgsCountErrorMessage, err.Error())
	})

	t.Run("should not return an error when there is at least one args provided", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{}
		localInstaller, err := installer.New()
		require.NoError(t, err)
		require.NotNil(t, localInstaller)

		cmd := newInstallCommand(localInstaller, dummyState)

		err = cmd.RunE(cmd, []string{"foo"})
		assert.NoError(t, err)

		err = cmd.RunE(cmd, []string{"foo", "bar"})
		assert.NoError(t, err)
	})

	t.Run("should return an error when binary already present in state", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		localInstaller, err := installer.New()
		require.NoError(t, err)
		require.NotNil(t, localInstaller)
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}
		name, version := "foo", "latest"
		binaryInfo := dto.BinaryInfo{
			FullName: name + "/" + name,
			Version:  version, Name: name, Owner: name, InstalledVersion: version,
		}

		dummyState.UpdateEntrie(binaryInfo)
		err = dummyState.Save()
		require.NoError(t, err)

		cmd := newInstallCommand(localInstaller, dummyState)
		err = cmd.RunE(cmd, []string{name})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "binary already installed")
	})
}

func TestInstallBinary(t *testing.T) {
	t.Run("should handle error on state failed", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{onError: true}
		localInstaller, err := installer.New()
		require.NoError(t, err)
		require.NotNil(t, localInstaller)
		cfg := InstallCommandConfig{
			azaInstaller: localInstaller,
			azaState:     dummyState,
		}
		binaryInfo := dto.BinaryInfo{
			FullName: TestBinaryFullName,
			Name:     TestBinaryName,
			Owner:    TestBinaryName,
			Version:  TestBinaryVersion,
		}

		err = installBinary(&binaryInfo, cfg)

		require.Error(t, err)
		assert.Equal(t, DummyStateErrorMessage, err.Error())
		assert.Equal(t, 0, dummyState.saveCount)
	})

	t.Run("should resolve url and install binary", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}
		dummyInstaller := &DummyInstaller{}
		dummyResolver := &DummyResolver{}
		cfg := InstallCommandConfig{
			azaInstaller: dummyInstaller,
			azaState:     dummyState,
		}
		binaryInfo := dto.BinaryInfo{
			FullName: TestBinaryFullName,
			Name:     TestBinaryName,
			Owner:    TestBinaryName,
			Version:  TestBinaryVersion,
		}
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := installBinary(&binaryInfo, cfg)
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		assert.NoError(t, err)
		assert.Equal(t, 1, dummyResolver.resolveCount,
			"should have called resolve method once")
		assert.Equal(t, 1, dummyInstaller.installCount,
			"should have called install method once")
		assert.Equal(t, 1, dummyState.saveCount,
			"should have called save method once")
		assert.True(t, dummyState.Has(binaryInfo.FullName), "binary should be present in state")
	})

	t.Run("should handle error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{}
		dummyInstaller := &DummyInstaller{onError: true}
		dummyResolver := &DummyResolver{}
		cfg := InstallCommandConfig{
			azaInstaller: dummyInstaller,
			azaState:     dummyState,
		}
		binaryInfo := dto.BinaryInfo{
			FullName: TestBinaryFullName,
			Name:     TestBinaryName,
			Owner:    TestBinaryName,
			Version:  TestBinaryVersion,
		}
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := installBinary(&binaryInfo, cfg)
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		require.Error(t, err)
		assert.Equal(t, DummyInstallerErrorMessage, err.Error())
		assert.Equal(t, 1, dummyResolver.resolveCount,
			"should have called resolve method once")
		assert.Equal(t, 1, dummyInstaller.installCount,
			"should have called install method once")
		assert.Equal(t, 0, dummyState.saveCount,
			"should not have called save method")
		assert.False(t, dummyState.Has(binaryInfo.FullName),
			"binary should not be present in state")
	})
}

func TestCreateBinaryInfo(t *testing.T) {
	t.Run("should handle binary name with different format", func(t *testing.T) {
		testCases := []struct {
			name     string
			binary   string
			version  string
			expected struct {
				owner string
				name  string
			}
		}{
			{
				name:    "simple format",
				binary:  "foo",
				version: "latest",
				expected: struct {
					owner string
					name  string
				}{owner: "foo", name: "foo"},
			},
			{
				name:    "with owner format",
				binary:  "foo/bar",
				version: "0.0.1",
				expected: struct {
					owner string
					name  string
				}{owner: "foo", name: "bar"},
			},
			{
				name:    "wrong format",
				binary:  "foo/bar/toto",
				version: "0.1.2",
				expected: struct {
					owner string
					name  string
				}{owner: "foo", name: "bar"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				binaryInfo := createBinaryInfo(tc.binary, tc.version)
				assert.Equal(t, tc.expected.owner, binaryInfo.Owner)
				assert.Equal(t, tc.expected.name, binaryInfo.Name)
				assert.Equal(t, fmt.Sprintf("%s/%s",
					tc.expected.owner, tc.expected.name), binaryInfo.FullName)
				assert.Equal(t, tc.version, binaryInfo.Version)
			})
		}
	})
}

func TestBinariesInfoFromArgs(t *testing.T) {
	testCases := []struct {
		name     string
		binaries []string
		version  string
		expected struct {
			length   int
			fullName []string
			version  string
		}
	}{
		{
			name:     "empty",
			binaries: []string{},
			version:  "latest",
			expected: struct {
				length   int
				fullName []string
				version  string
			}{
				length:   0,
				fullName: []string{""},
				version:  "",
			},
		},
		{
			name:     "one binary full",
			binaries: []string{"foo/bar"},
			version:  "latest",
			expected: struct {
				length   int
				fullName []string
				version  string
			}{
				length:   1,
				fullName: []string{"foo/bar"},
				version:  "latest",
			},
		},
		{
			name:     "one binary",
			binaries: []string{"foo"},
			version:  "latest",
			expected: struct {
				length   int
				fullName []string
				version  string
			}{
				length:   1,
				fullName: []string{"foo/foo"},
				version:  "latest",
			},
		},
		{
			name:     "nultiples binaries",
			binaries: []string{"foo/bar", "totocli", "cli/cli"},
			version:  "1.2.3",
			expected: struct {
				length   int
				fullName []string
				version  string
			}{
				length:   3,
				fullName: []string{"foo/bar", "totocli/totocli", "cli/cli"},
				version:  "1.2.3",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			binariesInfo := binariesInfoFromArgs(tc.binaries, tc.version)
			assert.Len(t, binariesInfo, tc.expected.length)
			for i, binaryInfo := range binariesInfo {
				assert.Equal(t, tc.expected.fullName[i], binaryInfo.FullName)
				assert.Equal(t, tc.expected.version, binaryInfo.Version)
			}
		})
	}
}
