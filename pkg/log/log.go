package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log, _ = zap.NewDevelopment()
var Logger = log.Sugar()

func ConfigureLogger() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	log, _ := config.Build()
	Logger = log.Sugar()
}
