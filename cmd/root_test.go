package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ludovic-alarcon/azabox/internal/logging"
	"gitlab.com/ludovic-alarcon/azabox/internal/resolver"
)

func TestRootCmd(t *testing.T) {
	t.Run("should init logger in persistentPreRun", func(t *testing.T) {
		t.Cleanup(func() {
			logging.Logger = nil
			logging.LogLevel = ""
		})

		assert.Nil(t, logging.Logger)

		err := rootCmd.PersistentPreRunE(rootCmd.Root(), []string{})
		require.NoError(t, err)
		assert.NotNil(t, logging.Logger)
	})

	t.Run("should init registry resolver with default resolvers", func(t *testing.T) {
		t.Cleanup(func() {
			logging.Logger = nil
			logging.LogLevel = ""
		})
		err := rootCmd.Execute()
		assert.NoError(t, err)
		resolvers := resolver.GetRegistryResolver().GetResolvers()
		assert.Len(t, resolvers, 1)
	})

	t.Run("should set logLevel when flag is used", func(t *testing.T) {
		t.Cleanup(func() {
			logging.Logger = nil
			logging.LogLevel = ""
		})

		rootCmd.SetArgs([]string{"--log-level", "debug"})
		err := rootCmd.Execute()
		assert.NoError(t, err)
		assert.Equal(t, "debug", logging.LogLevel)
	})

	t.Run("should handle error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.LogLevel = ""
			logging.Logger = nil
		})

		expectedContains := "some error"
		saveRootCmd := rootCmd
		defer func() { rootCmd = saveRootCmd }()

		fakeCmd := &cobra.Command{
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New(expectedContains)
			},
		}

		rootCmd = fakeCmd
		err := logging.InitLogger(logging.Config{Encoding: logging.Json})
		require.NoError(t, err)

		// capture stderr
		saveSterr := os.Stderr
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stderr = w

		defer func() {
			os.Stderr = saveSterr
			w.Close()
			r.Close()
		}()

		err = Execute()
		w.Close()
		os.Stderr = saveSterr
		assert.Error(t, err)

		// read captured output
		var buff bytes.Buffer
		_, err = io.Copy(&buff, r)
		assert.NoError(t, err)
		r.Close()

		output := buff.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, expectedContains)
	})

	t.Run("should execute", func(t *testing.T) {
		rootCmd.SetArgs([]string{"list"})
		err := Execute()
		assert.NoError(t, err)
	})
}
