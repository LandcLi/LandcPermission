package logger

import (
	"github.com/LandcLi/landc-go/log/facade"
)

var globalLogger facade.Logger

// InitLogger initializes the logger with LogConfig.
func InitLogger(config *facade.LogConfig) {
	if globalLogger == nil {
		globalLogger = facade.GetLoggerWithLogConfig(config)
	}
}

// InitLoggerWithLogger initializes with an existing logger.
func InitLoggerWithLogger(logger facade.Logger) {
	globalLogger = logger
}

// GetLogger returns the logger instance.
func GetLogger() facade.Logger {
	return globalLogger
}
