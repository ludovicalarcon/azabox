package cmd

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/dto"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
)

const FakeVersionToUpdate = "X.Y.Z"

func TestResolveLatestVersion(t *testing.T) {
	t.Run("should resolve latest version and return proper resolver", func(t *testing.T) {
		resolver.GetRegistryResolver().Register(&DummyResolver{})
		binaryInfo := dto.BinaryInfo{
			FullName: TestBinaryFullName,
			Owner:    TestBinaryName,
			Name:     TestBinaryName,
			Resolver: DummyResolverName,
		}

		version, lresolver, err := resolveLatestVersion(binaryInfo)

		require.NoError(t, err)
		require.NotNil(t, lresolver)
		assert.IsType(t, &DummyResolver{}, lresolver)
		assert.NotEmpty(t, version)
		assert.Equal(t, version, TestBinaryVersion)

		resolver.GetRegistryResolver().Unregister(lresolver)
	})

	t.Run("should handle resolver not found", func(t *testing.T) {
		resolverName := "foobar"
		binaryInfo := dto.BinaryInfo{
			FullName: TestBinaryFullName,
			Owner:    TestBinaryName,
			Name:     TestBinaryName,
			Resolver: resolverName,
		}

		version, lresolver, err := resolveLatestVersion(binaryInfo)

		require.Error(t, err)
		assert.Contains(t, err.Error(), resolverName)
		assert.Nil(t, lresolver)
		assert.Empty(t, version)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("should update binary", func(t *testing.T) {
		version := FakeVersionToUpdate
		dummyState := DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}
		binaryInfo := dto.BinaryInfo{
			FullName:         TestBinaryName,
			Owner:            TestBinaryName,
			Name:             TestBinaryName,
			Version:          version,
			InstalledVersion: version,
			Resolver:         DummyResolverName,
		}
		dummyState.UpdateEntrie(binaryInfo)
		dummyResolver := DummyResolver{onError: false}
		dummyInstaller := DummyInstaller{onError: false}
		cfg := UpdateCommandConfig{
			azaInstaller: &dummyInstaller,
			azaState:     &dummyState,
		}

		err := update(&dummyResolver, &binaryInfo, cfg)

		assert.NoError(t, err)
		assert.Equal(t, 1, dummyResolver.resolveCount, "resolve method should have been called once")
		assert.Equal(t, 1, dummyInstaller.installCount, "install method should have been called once")
		assert.Len(t, cfg.azaState.Entries(), 1)

		info, ok := cfg.azaState.Entry(TestBinaryName)
		assert.True(t, ok)
		assert.Equal(t, TestBinaryVersion, info.InstalledVersion)
	})

	t.Run("should handle error", func(t *testing.T) {
		testCases := []struct {
			name                 string
			onResolveError       bool
			onInstallError       bool
			installCount         int
			expectedErrorMessage string
		}{
			{
				name:                 "resolve return an error",
				onResolveError:       true,
				onInstallError:       false,
				installCount:         0,
				expectedErrorMessage: DummyResolverErrorMessage,
			}, {
				name:                 "install return error",
				onResolveError:       false,
				onInstallError:       true,
				installCount:         1,
				expectedErrorMessage: DummyInstallerErrorMessage,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				version := FakeVersionToUpdate
				dummyState := DummyState{
					binaries: make(map[string]dto.BinaryInfo, 1),
				}
				binaryInfo := dto.BinaryInfo{
					FullName:         TestBinaryName,
					Owner:            TestBinaryName,
					Name:             TestBinaryName,
					Version:          version,
					InstalledVersion: version,
					Resolver:         DummyResolverName,
				}
				dummyState.UpdateEntrie(binaryInfo)
				dummyResolver := DummyResolver{onError: tc.onResolveError}
				dummyInstaller := DummyInstaller{onError: tc.onInstallError}
				cfg := UpdateCommandConfig{
					azaInstaller: &dummyInstaller,
					azaState:     &dummyState,
				}

				err := update(&dummyResolver, &binaryInfo, cfg)

				require.Error(t, err)
				assert.Equal(t, tc.installCount, dummyInstaller.installCount)
				assert.Equal(t, 1, dummyResolver.resolveCount)
				assert.Equal(t, tc.expectedErrorMessage, err.Error())
			})
		}
	})
}

func TestCheckUpdate(t *testing.T) {
	t.Run("should check for update", func(t *testing.T) {
		testCases := []struct {
			name             string
			onResolveError   bool
			onInstallError   bool
			installCount     int
			installedVersion string
		}{
			{
				name:             "should update as newer version available",
				onResolveError:   false,
				onInstallError:   false,
				installCount:     1,
				installedVersion: FakeVersionToUpdate,
			}, {
				name:             "should not update as version is the same",
				onResolveError:   false,
				onInstallError:   false,
				installCount:     0,
				installedVersion: TestBinaryVersion,
			}, {
				name:             "should handle error in resolver",
				onResolveError:   true,
				onInstallError:   false,
				installCount:     0,
				installedVersion: FakeVersionToUpdate,
			}, {
				name:             "should handle error in installer",
				onResolveError:   false,
				onInstallError:   true,
				installCount:     1,
				installedVersion: FakeVersionToUpdate,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Cleanup(func() {
					logging.LogLevel = ""
					logging.Logger = nil
				})

				initLoggerForTest()
				dummyState := DummyState{
					binaries: make(map[string]dto.BinaryInfo, 1),
				}
				binaryInfo := dto.BinaryInfo{
					FullName:         TestBinaryName,
					Owner:            TestBinaryName,
					Name:             TestBinaryName,
					Version:          tc.installedVersion,
					InstalledVersion: tc.installedVersion,
					Resolver:         DummyResolverName,
				}
				dummyState.UpdateEntrie(binaryInfo)
				dummyResolver := DummyResolver{onError: tc.onResolveError}
				dummyInstaller := DummyInstaller{onError: tc.onInstallError}
				cfg := UpdateCommandConfig{
					azaInstaller: &dummyInstaller,
					azaState:     &dummyState,
				}
				resolver.GetRegistryResolver().GetResolvers().Clear()
				resolver.GetRegistryResolver().Register(&dummyResolver)

				err := checkUpdate(binaryInfo, cfg)
				resolver.GetRegistryResolver().Unregister(&dummyResolver)

				switch {
				case tc.onResolveError:
					require.Error(t, err)
					assert.Equal(t, DummyResolverErrorMessage, err.Error())
				case tc.onInstallError:
					require.Error(t, err)
					assert.Equal(t, DummyInstallerErrorMessage, err.Error())
				default:
					info, ok := cfg.azaState.Entry(TestBinaryName)
					require.NoError(t, err)
					assert.True(t, ok, "binary should be in state")
					assert.Equal(t, TestBinaryVersion, info.InstalledVersion)
				}
				assert.Equal(t, tc.installCount, dummyInstaller.installCount)
			})
		}
	})
}

func TestExecuteUpdateCommand(t *testing.T) {
	t.Run("should handle list of binary to update", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 3),
		}

		binaryName1 := fmt.Sprintf("%s1", TestBinaryName)
		binaryName2 := fmt.Sprintf("%s2", TestBinaryName)

		for i := 1; i < 3; i++ {
			binaryInfo := dto.BinaryInfo{
				Version:          FakeVersionToUpdate,
				InstalledVersion: FakeVersionToUpdate,
				Resolver:         DummyResolverName,
			}
			binaryInfo.Name = fmt.Sprintf("%s%d", TestBinaryName, i)
			binaryInfo.Owner = fmt.Sprintf("%s%d", TestBinaryName, i)
			binaryInfo.FullName = fmt.Sprintf("%s/%s", binaryInfo.Owner, binaryInfo.Name)
			dummyState.UpdateEntrie(binaryInfo)
		}

		dummyResolver := &DummyResolver{}
		dummyInstaller := &DummyInstaller{}
		cfg := UpdateCommandConfig{
			azaInstaller: dummyInstaller,
			azaState:     dummyState,
		}
		resolver.GetRegistryResolver().GetResolvers().Clear()
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := executeUpdateCommand(cfg, binaryName1, binaryName2)
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		assert.NoError(t, err)
		assert.Equal(t, 2, dummyInstaller.installCount)
		assert.Equal(t, 1, dummyState.saveCount)
		info, _ := dummyState.Entry(fmt.Sprintf("%s/%s", binaryName1, binaryName1))
		assert.Equal(t, TestBinaryVersion, info.InstalledVersion)
		info, _ = dummyState.Entry(fmt.Sprintf("%s/%s", binaryName2, binaryName2))
		assert.Equal(t, TestBinaryVersion, info.InstalledVersion)
	})

	t.Run("should handle error on state", func(t *testing.T) {
		dummyState := &DummyState{onError: true}
		dummyInstaller := &DummyInstaller{}
		cfg := UpdateCommandConfig{
			azaState:     dummyState,
			azaInstaller: dummyInstaller,
		}

		err := executeUpdateCommand(cfg, "")
		require.Error(t, err)
		assert.Equal(t, DummyStateErrorMessage, err.Error())
		assert.Equal(t, 0, dummyState.saveCount, "state save method should not be called")
	})

	t.Run("should handle binary not in state", func(t *testing.T) {
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}
		dummyInstaller := &DummyInstaller{}
		cfg := UpdateCommandConfig{
			azaState:     dummyState,
			azaInstaller: dummyInstaller,
		}

		err := executeUpdateCommand(cfg, "unknown")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "is not installed (or not managed by azabox)")
		assert.Equal(t, 0, dummyState.saveCount, "state save method should not be called")
	})

	t.Run("should handle error in the update process", func(t *testing.T) {
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}
		dummyInstaller := &DummyInstaller{}
		dummyResolver := &DummyResolver{onError: true}
		binaryInfo := dto.BinaryInfo{
			FullName:         TestBinaryFullName,
			Name:             TestBinaryName,
			Owner:            TestBinaryName,
			Version:          TestBinaryVersion,
			InstalledVersion: TestBinaryVersion,
			Resolver:         DummyResolverName,
		}
		cfg := UpdateCommandConfig{
			azaState:     dummyState,
			azaInstaller: dummyInstaller,
		}

		resolver.GetRegistryResolver().GetResolvers().Clear()
		resolver.GetRegistryResolver().Register(dummyResolver)
		dummyState.UpdateEntrie(binaryInfo)

		err := executeUpdateCommand(cfg, TestBinaryName)
		require.Error(t, err)
		assert.Equal(t, DummyResolverErrorMessage, err.Error())
		assert.Equal(t, 0, dummyState.saveCount, "state save method should not be called")
	})

	t.Run("should update all binaries when no parameter are given", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 3),
		}

		for i := 1; i < 4; i++ {
			binaryInfo := dto.BinaryInfo{
				Version:          FakeVersionToUpdate,
				InstalledVersion: FakeVersionToUpdate,
				Resolver:         DummyResolverName,
			}
			binaryInfo.Name = fmt.Sprintf("%s%d", TestBinaryName, i)
			binaryInfo.Owner = fmt.Sprintf("%s%d", TestBinaryName, i)
			binaryInfo.FullName = fmt.Sprintf("%s/%s", binaryInfo.Owner, binaryInfo.Name)
			dummyState.UpdateEntrie(binaryInfo)
		}

		dummyResolver := &DummyResolver{}
		dummyInstaller := &DummyInstaller{}
		cfg := UpdateCommandConfig{
			azaInstaller: dummyInstaller,
			azaState:     dummyState,
		}
		resolver.GetRegistryResolver().GetResolvers().Clear()
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := executeUpdateCommand(cfg, []string{}...)
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		assert.NoError(t, err)
		assert.Equal(t, 3, dummyInstaller.installCount, "install method should have been called")
		assert.Equal(t, 1, dummyState.saveCount, "state save method should have been called once")
	})

	t.Run("should handle error on update when no parameter are given", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}

		binaryInfo := dto.BinaryInfo{
			FullName:         TestBinaryFullName,
			Name:             TestBinaryName,
			Owner:            TestBinaryName,
			Version:          FakeVersionToUpdate,
			InstalledVersion: FakeVersionToUpdate,
			Resolver:         DummyResolverName,
		}
		dummyState.UpdateEntrie(binaryInfo)

		dummyResolver := &DummyResolver{}
		dummyInstaller := &DummyInstaller{onError: true}
		cfg := UpdateCommandConfig{
			azaInstaller: dummyInstaller,
			azaState:     dummyState,
		}
		resolver.GetRegistryResolver().GetResolvers().Clear()
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := executeUpdateCommand(cfg, []string{}...)
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		assert.Error(t, err)
		assert.Equal(t, DummyInstallerErrorMessage, err.Error())
		assert.Equal(t, 0, dummyState.saveCount, "state save method should not be called")
	})
}

func TestNewUpdateCommand(t *testing.T) {
	t.Run("should return any exec error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		initLoggerForTest()
		dummyState := &DummyState{
			binaries: make(map[string]dto.BinaryInfo, 1),
		}

		binaryInfo := dto.BinaryInfo{
			FullName:         TestBinaryFullName,
			Name:             TestBinaryName,
			Owner:            TestBinaryName,
			Version:          FakeVersionToUpdate,
			InstalledVersion: FakeVersionToUpdate,
			Resolver:         DummyResolverName,
		}
		dummyState.UpdateEntrie(binaryInfo)

		dummyResolver := &DummyResolver{onError: true}
		dummyInstaller := &DummyInstaller{}
		resolver.GetRegistryResolver().GetResolvers().Clear()
		resolver.GetRegistryResolver().Register(dummyResolver)

		err := newUpdateCommand(dummyInstaller, dummyState).Execute()
		resolver.GetRegistryResolver().Unregister(dummyResolver)

		assert.Error(t, err)
		assert.Equal(t, DummyResolverErrorMessage, err.Error())
		assert.Equal(t, 0, dummyState.saveCount, "state save method should not be called")
	})
}
