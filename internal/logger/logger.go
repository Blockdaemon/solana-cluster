package logger

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Flags = pflag.NewFlagSet("logger", pflag.ExitOnError)

	logLevel  = LogLevel{zap.InfoLevel}
	logFormat string
)

func init() {
	Flags.Var(&logLevel, "log-level", "Log level")
	Flags.StringVar(&logFormat, "log-format", "console", "Log format (console, json)")
}

func GetLogger() *zap.Logger {
	var config zap.Config
	if logFormat == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.DisableStacktrace = true
	}
	config.DisableCaller = true
	config.Level.SetLevel(logLevel.Level)
	logger, err := config.Build()
	if err != nil {
		panic(err.Error())
	}
	return logger
}

// LogLevel is required to use zap level as a pflag.
type LogLevel struct{ zapcore.Level }

func (LogLevel) Type() string {
	return "string"
}
