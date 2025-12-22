package logging

import (
	"os"
	"sync"

	azalogger "gitlab.com/ludovic-alarcon/aza-logger"
)

const (
	logLevelEnvVar           = "AZABOX_LOG_LEVEL"
	panicNotInitializedError = "logger must be initialized"
)

var (
	LogLevel string
	logger   azalogger.Logger
	once     sync.Once
)

func setLogLevel(flagLogLevel string) {
	logLevel := flagLogLevel
	if logLevel == "" {
		logLevel = os.Getenv(logLevelEnvVar)
		if logLevel == "" {
			logLevel = "info"
		}
	}
	os.Setenv(azalogger.LogLevelEnvVar, logLevel)
}

func InitLogger() error {
	var err error
	once.Do(func() {
		setLogLevel(LogLevel)
		logger, err = azalogger.NewLogger(azalogger.Config{
			Backend:  azalogger.ZapBackend,
			Env:      azalogger.ProdEnvironment,
			LogLevel: azalogger.InfoLevel,
		})
		if err == nil {
			logger = logger.With("cli", "azalogger")
		}
	})
	return err
}

func Logger() azalogger.Logger {
	if logger == nil {
		panic(panicNotInitializedError)
	}
	return logger
}

// Helpers shared for tests in all packages
func UseInMemoryLogger() {
	cfg := azalogger.Config{
		Backend:  azalogger.InMemoryBackend,
		Env:      azalogger.ProdEnvironment,
		LogLevel: azalogger.DebugLevel,
	}
	logger = azalogger.NewInMemoryLogger(cfg)
}
