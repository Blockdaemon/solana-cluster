// Copyright 2022 Blockdaemon Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func GetConsoleLogger() *zap.Logger {
	logConfig := zap.Config{
		Level:             zap.AtomicLevel{},
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
	}
	logConfig.Level.SetLevel(zap.InfoLevel)
	log, err := logConfig.Build()
	if err != nil {
		panic(err.Error())
	}
	return log
}
