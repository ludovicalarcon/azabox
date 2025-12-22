package logging

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	azalogger "gitlab.com/ludovic-alarcon/aza-logger"
)

func TestSetLogLevel(t *testing.T) {
	t.Run("should init logger level with env var when set and flag not used", func(t *testing.T) {
		t.Cleanup(func() {
			_ = os.Unsetenv(azalogger.LogLevelEnvVar)
		})

		expected := "warn"
		t.Setenv(logLevelEnvVar, expected)
		setLogLevel("")

		assert.Equal(t, expected, os.Getenv(azalogger.LogLevelEnvVar))
	})

	t.Run("should init logger level with flag", func(t *testing.T) {
		t.Cleanup(func() {
			_ = os.Unsetenv(azalogger.LogLevelEnvVar)
		})

		expected := "debug"
		t.Setenv(logLevelEnvVar, expected)
		setLogLevel(expected)

		assert.Equal(t, expected, os.Getenv(azalogger.LogLevelEnvVar))
	})

	t.Run("should init logger level with info when nothing is set", func(t *testing.T) {
		t.Cleanup(func() {
			_ = os.Unsetenv(azalogger.LogLevelEnvVar)
		})

		expected := "info"
		setLogLevel("")

		assert.Equal(t, expected, os.Getenv(azalogger.LogLevelEnvVar))
	})
}

func TestInitLogger(t *testing.T) {
	t.Run("should init logger only once", func(t *testing.T) {
		t.Cleanup(func() {
			LogLevel = ""
			_ = os.Unsetenv(azalogger.LogLevelEnvVar)
			logger = nil
		})

		expectedLogLevel := "debug"
		LogLevel = expectedLogLevel
		err := InitLogger()
		require.Nil(t, err, "unexpected error on first InitLogger call")
		firstLogger := logger

		LogLevel = "warn"
		err = InitLogger()
		require.Nil(t, err, "unexpected error on second InitLogger call")
		secondLogger := logger

		retLoggger := Logger()

		assert.Equal(t, firstLogger, secondLogger, "logger was reinitialized...")
		assert.Equal(t, firstLogger, retLoggger, "logger was reinitialized...")
		assert.Equal(t, expectedLogLevel, os.Getenv(azalogger.LogLevelEnvVar))
		assert.Equal(t, expectedLogLevel, secondLogger.LogLevel())
	})
}

func TestLogger(t *testing.T) {
	t.Run("should panic if logger is not initialized", func(t *testing.T) {
		defer func() {
			r := recover()
			require.NotNil(t, r, "expected panic")
			assert.Equal(t, panicNotInitializedError, r)
		}()

		_ = Logger()
	})
}

func TestUseInMemoryLogger(t *testing.T) {
	UseInMemoryLogger()
	assert.IsType(t, &azalogger.InMemoryLogger{}, logger)
	assert.Equal(t, azalogger.DebugLevel.String(), logger.LogLevel())
}
