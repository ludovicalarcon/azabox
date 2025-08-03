package logging

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetEncoding(t *testing.T) {
	t.Run("should return proper encoding", func(t *testing.T) {
		got := getEndoding(Console)
		assert.Equal(t, "console", got)

		got = getEndoding(Json)
		assert.Equal(t, "json", got)
	})

	t.Run("should default to json", func(t *testing.T) {
		unknownEncoding := 5
		got := getEndoding(Encoding(unknownEncoding))
		assert.Equal(t, "json", got)
	})
}

func TestCreateZapConfig(t *testing.T) {
	t.Run("should create default config", func(t *testing.T) {
		expected := zap.NewProductionEncoderConfig()

		got := createZapConfig(Config{Encoding: Json})
		assert.Equal(t, "info", got.Level.String())
		assert.Equal(t, "json", got.Encoding)
		require.Len(t, got.OutputPaths, 1)
		assert.Equal(t, "stdout", got.OutputPaths[0])
		assert.Equal(t, expected.MessageKey, got.EncoderConfig.MessageKey)
		assert.Equal(t, expected.LevelKey, got.EncoderConfig.LevelKey)
		assert.Equal(t, expected.TimeKey, got.EncoderConfig.TimeKey)
		assert.Equal(t, expected.NameKey, got.EncoderConfig.NameKey)
		assert.Equal(t, expected.CallerKey, got.EncoderConfig.CallerKey)
		assert.Equal(t, expected.FunctionKey, got.EncoderConfig.FunctionKey)
		assert.Equal(t, expected.StacktraceKey, got.EncoderConfig.StacktraceKey)
	})

	t.Run("should override log level from env var", func(t *testing.T) {
		t.Setenv(logLevelEnvVar, "warn")
		got := createZapConfig(Config{Encoding: Console})
		assert.Equal(t, "warn", got.Level.String())
	})

	t.Run("should override log level from log-level flag even if env var is set", func(t *testing.T) {
		t.Cleanup(func() {
			LogLevel = ""
			Logger = nil
		})

		t.Setenv(logLevelEnvVar, "debug")
		LogLevel = "warn"
		got := createZapConfig(Config{Encoding: Console})
		assert.Equal(t, "warn", got.Level.String())
	})

	t.Run("should default log level to info when unknow log-level is passed", func(t *testing.T) {
		t.Cleanup(func() {
			LogLevel = ""
			Logger = nil
		})

		t.Setenv(logLevelEnvVar, "foo")
		got := createZapConfig(Config{Encoding: Console})
		assert.Equal(t, "info", got.Level.String())

		LogLevel = "bar"
		got = createZapConfig(Config{Encoding: Console})
		assert.Equal(t, "info", got.Level.String())
	})

	t.Run("should use dev encoder config when encoding is console", func(t *testing.T) {
		expected := zap.NewDevelopmentEncoderConfig()

		got := createZapConfig(Config{Encoding: Console})
		assert.Equal(t, expected.MessageKey, got.EncoderConfig.MessageKey)
		assert.Equal(t, expected.LevelKey, got.EncoderConfig.LevelKey)
		assert.Equal(t, expected.TimeKey, got.EncoderConfig.TimeKey)
		assert.Equal(t, expected.NameKey, got.EncoderConfig.NameKey)
		assert.Equal(t, expected.CallerKey, got.EncoderConfig.CallerKey)
		assert.Equal(t, expected.FunctionKey, got.EncoderConfig.FunctionKey)
		assert.Equal(t, expected.StacktraceKey, got.EncoderConfig.StacktraceKey)
	})

	t.Run("should activate stacktrace and caller only in debug log-level", func(t *testing.T) {
		t.Cleanup(func() {
			LogLevel = ""
		})

		testcases := []struct {
			name     string
			disabled bool
			logLevel string
		}{
			{
				name:     "logLevel debug should have stacktrace",
				disabled: false,
				logLevel: "debug",
			},
			{
				name:     "logLevel info should not have stacktrace",
				disabled: true,
				logLevel: "info",
			},
			{
				name:     "logLevel warn should not have stacktrace",
				disabled: true,
				logLevel: "warn",
			},
			{
				name:     "logLevel warn should not have stacktrace",
				disabled: true,
				logLevel: "error",
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				LogLevel = tc.logLevel
				cfg := createZapConfig(Config{Encoding: Console})
				assert.Equal(t, tc.disabled, cfg.DisableStacktrace)
				assert.Equal(t, tc.disabled, cfg.DisableCaller)
			})
		}
	})
}

func TestInitLogger(t *testing.T) {
	t.Run("should init the logger", func(t *testing.T) {
		t.Cleanup(func() {
			LogLevel = ""
			Logger = nil
		})

		Logger = nil
		LogLevel = "debug"
		expectLogMessage := "test message log"

		// capture stdout
		saveStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		defer func() {
			os.Stdout = saveStdout
			w.Close()
			r.Close()
		}()

		err = InitLogger(Config{Console})
		require.NoError(t, err)
		require.NotNil(t, Logger)
		require.NotNil(t, Logger.Desugar())

		Logger.Debug(expectLogMessage)
		_ = Logger.Sync()
		w.Close()
		os.Stdout = saveStdout

		// read captured output
		var buff bytes.Buffer
		_, err = io.Copy(&buff, r)
		require.NoError(t, err)
		r.Close()

		output := buff.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, expectLogMessage)
	})
}
