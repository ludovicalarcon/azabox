package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger   *zap.SugaredLogger
	LogLevel string
)

type Encoding int

const (
	Console Encoding = iota
	Json
)

type Config struct {
	Encoding Encoding
}

func getEndoding(enc Encoding) string {
	switch enc {
	case Console:
		return "console"
	case Json:
		return "json"
	}
	return "json"
}

func createZapConfig(cfg Config) zap.Config {
	level := LogLevel
	if level == "" {
		level = os.Getenv("AZABOX_LOG_LEVEL")
		if level == "" {
			level = "info"
		}
	}
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	cfgZap := zap.Config{
		Level:         zap.NewAtomicLevelAt(zapLevel),
		Encoding:      getEndoding(cfg.Encoding),
		OutputPaths:   []string{"stdout"},
		EncoderConfig: zap.NewProductionEncoderConfig(),
	}

	if cfg.Encoding == Console {
		cfgZap.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	return cfgZap
}

func InitLogger(cfg Config) error {
	cfgZap := createZapConfig(cfg)
	zapLogger, err := cfgZap.Build()
	if err != nil {
		return err
	}
	Logger = zapLogger.Sugar()
	return nil
}
