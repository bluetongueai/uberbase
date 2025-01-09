package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

// LogLevel represents the logging level
type LogLevel string

const (
	// DebugLevel logs debug or higher
	DebugLevel LogLevel = "debug"
	// InfoLevel logs info or higher
	InfoLevel LogLevel = "info"
	// WarnLevel logs warn or higher
	WarnLevel LogLevel = "warn"
	// ErrorLevel logs error or higher
	ErrorLevel LogLevel = "error"
)

// InitLogging initializes the logger with default settings
func InitLogging() {
	// Set default level to info
	Logger.SetLevel(logrus.InfoLevel)

	// Configure formatter
	Logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})

	// Set output to stderr for better handling in containers
	Logger.SetOutput(os.Stderr)
}

// SetLevel sets the logging level
func SetLevel(level LogLevel) {
	switch level {
	case DebugLevel:
		Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
}

// SetOutput sets the output writer for the logger
func SetOutput(w io.Writer) {
	Logger.SetOutput(w)
}

// SetDebugLevel is a convenience function to set debug level
func SetDebugLevel() {
	SetLevel(DebugLevel)
}

// SetInfoLevel is a convenience function to set info level
func SetInfoLevel() {
	SetLevel(InfoLevel)
}
