package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// LevelHook implements logrus.Hook interface to write certain levels to different outputs
type LevelHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

func (hook *LevelHook) Fire(entry *logrus.Entry) error {
	writer, ok := hook.Writer.(*FormattedWriter)
	if !ok {
		return fmt.Errorf("writer is not a FormattedWriter")
	}

	// Format the entry using the writer's specific formatter
	bytes, err := writer.Formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = writer.Writer.Write(bytes)
	return err
}

func (hook *LevelHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// CustomTextFormatter removes level prefix for stdout
type CustomTextFormatter struct {
	logrus.TextFormatter
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Just return the message with a newline
	return []byte(entry.Message + "\n"), nil
}

// InitLogging initializes the logger with default settings
func InitLogging() error {
	// Set default level to debug to capture all logs
	Logger.SetLevel(logrus.DebugLevel)

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file
	logFile, err := os.OpenFile(
		filepath.Join("logs", "debug.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Set default output to discard (we'll use hooks instead)
	Logger.SetOutput(io.Discard)

	// Add hook for info and above to stdout with minimal formatting
	Logger.AddHook(&LevelHook{
		Writer: &FormattedWriter{
			Writer: os.Stdout,
			Formatter: &CustomTextFormatter{
				TextFormatter: logrus.TextFormatter{
					DisableTimestamp: true,
					DisableQuote:     true,
					ForceColors:      true,
				},
			},
		},
		LogLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})

	// Add hook for all levels to file with detailed formatting
	Logger.AddHook(&LevelHook{
		Writer: &FormattedWriter{
			Writer: logFile,
			Formatter: &logrus.TextFormatter{
				DisableColors:          true,
				FullTimestamp:          true,
				TimestampFormat:        "2006-01-02 15:04:05",
				DisableQuote:           true,
				DisableLevelTruncation: true,
				PadLevelText:           true,
			},
		},
		LogLevels: []logrus.Level{
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
			logrus.PanicLevel,
		},
	})

	return nil
}

// FormattedWriter wraps an io.Writer with a specific formatter
type FormattedWriter struct {
	Writer    io.Writer
	Formatter logrus.Formatter
}

func (fw *FormattedWriter) Write(p []byte) (n int, err error) {
	return fw.Writer.Write(p)
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

// LogKeyValues logs a message followed by aligned key-value pairs
func LogKeyValues(msg string, kvs map[string]string) {
	var maxKeyLen int
	for k := range kvs {
		if len(k) > maxKeyLen {
			maxKeyLen = len(k)
		}
	}

	Logger.Info(msg)
	for k, v := range kvs {
		padding := strings.Repeat(" ", maxKeyLen-len(k))
		Logger.Infof("  %s%s: %s", k, padding, v)
	}
}
