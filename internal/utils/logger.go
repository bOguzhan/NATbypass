// internal/utils/logger.go
package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger with some additional convenience methods
type Logger struct {
	*logrus.Logger
	component string
}

// NewLogger creates a new logger with the specified component name and log level
func NewLogger(component string, level string) *Logger {
	logger := logrus.New()

	// Set output format
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	})

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Set log level from string
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		// Default to info level if parsing fails
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return &Logger{
		Logger:    logger,
		component: component,
	}
}

// WithFields returns a logrus entry with fields prefilled with component
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	// Always include component field
	if fields == nil {
		fields = logrus.Fields{}
	}
	fields["component"] = l.component
	return l.Logger.WithFields(fields)
}

// Trace logs a message at level Trace
func (l *Logger) Trace(args ...interface{}) {
	l.WithFields(nil).Trace(args...)
}

// Debug logs a message at level Debug
func (l *Logger) Debug(args ...interface{}) {
	l.WithFields(nil).Debug(args...)
}

// Info logs a message at level Info
func (l *Logger) Info(args ...interface{}) {
	l.WithFields(nil).Info(args...)
}

// Warn logs a message at level Warn
func (l *Logger) Warn(args ...interface{}) {
	l.WithFields(nil).Warn(args...)
}

// Error logs a message at level Error
func (l *Logger) Error(args ...interface{}) {
	l.WithFields(nil).Error(args...)
}

// Fatal logs a message at level Fatal then the process will exit with status set to 1
func (l *Logger) Fatal(args ...interface{}) {
	l.WithFields(nil).Fatal(args...)
}

// Tracef logs a formatted message at level Trace
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.WithFields(nil).Tracef(format, args...)
}

// Debugf logs a formatted message at level Debug
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.WithFields(nil).Debugf(format, args...)
}

// Infof logs a formatted message at level Info
func (l *Logger) Infof(format string, args ...interface{}) {
	l.WithFields(nil).Infof(format, args...)
}

// Warnf logs a formatted message at level Warn
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.WithFields(nil).Warnf(format, args...)
}

// Errorf logs a formatted message at level Error
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.WithFields(nil).Errorf(format, args...)
}

// Fatalf logs a formatted message at level Fatal then the process will exit with status set to 1
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.WithFields(nil).Fatalf(format, args...)
}
