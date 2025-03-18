package ibc

import (
	"fmt"
	"testing"
	"time"
)

// LogLevel represents different levels of logging
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

// timestampedLog logs a message with a timestamp
// This function is used across multiple test files to provide consistent
// timestamp format and make key events easier to track
func timestampedLog(t *testing.T, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	t.Logf("[%s] %s", timestamp, message)
}

// logWithLevel logs a message with a timestamp and a specific log level
func logWithLevel(t *testing.T, level LogLevel, format string, args ...interface{}) {
	var levelPrefix string
	switch level {
	case LogLevelDebug:
		levelPrefix = "DEBUG"
	case LogLevelInfo:
		levelPrefix = "INFO"
	case LogLevelWarning:
		levelPrefix = "WARN"
	case LogLevelError:
		levelPrefix = "ERROR"
	}

	message := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	t.Logf("[%s] [%s] %s", timestamp, levelPrefix, message)
}

// debugLog logs a debug message
func debugLog(t *testing.T, format string, args ...interface{}) {
	logWithLevel(t, LogLevelDebug, format, args...)
}

// infoLog logs an info message
func infoLog(t *testing.T, format string, args ...interface{}) {
	logWithLevel(t, LogLevelInfo, format, args...)
}

// warnLog logs a warning message
func warnLog(t *testing.T, format string, args ...interface{}) {
	logWithLevel(t, LogLevelWarning, format, args...)
}

// errorLog logs an error message
func errorLog(t *testing.T, format string, args ...interface{}) {
	logWithLevel(t, LogLevelError, format, args...)
}
