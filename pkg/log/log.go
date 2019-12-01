package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log, _ = zap.NewDevelopment()
var Logger = log.Sugar()

func ConfigureLogger(debugMode bool) {
	config := zap.NewDevelopmentConfig()
	if !debugMode {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.Development = false
	}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log, _ := config.Build()
	Logger = log.Sugar()
}
