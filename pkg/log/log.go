package log

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func InitLogger() {
	var zapLogger, _ = zap.NewDevelopment()
	Logger = zapLogger.Sugar()
	Logger.Debug("ho")
}
