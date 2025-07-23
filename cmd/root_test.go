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
)

func TestRootCmd(t *testing.T) {
	t.Run("should show help without error", func(t *testing.T) {
		t.Cleanup(func() {
			logging.Logger = nil
			logging.LogLevel = ""
		})

		buf := new(bytes.Buffer)
		rootCmd.SetOut(buf)
		rootCmd.SetArgs([]string{})

		Execute()

		output := buf.String()
		assert.Contains(t, output, "azabox is a CLI tool")
	})

	t.Run("should init logger in persistentPreRun", func(t *testing.T) {
		t.Cleanup(func() {
			logging.Logger = nil
			logging.LogLevel = ""
		})

		assert.Nil(t, logging.Logger)

		rootCmd.PersistentPreRun(rootCmd.Root(), []string{})
		assert.NotNil(t, logging.Logger)
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
		logging.InitLogger(logging.Config{Encoding: logging.Json})

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
		io.Copy(&buff, r)
		r.Close()

		output := buff.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, expectedContains)
	})
}
