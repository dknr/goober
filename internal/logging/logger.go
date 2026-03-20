package logging

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

func NewLogger(level string) *Logger {
	// Default to info level
	_ = level // TODO: Implement log levels

	return &Logger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *Logger) Info(msg string) {
	l.logger.Printf("[INFO] %s", msg)
}

func (l *Logger) Infof(msg string, args ...interface{}) {
	l.logger.Printf("[INFO] "+msg, args...)
}

func (l *Logger) Error(msg string) {
	l.logger.Printf("[ERROR] %s", msg)
}

func (l *Logger) Errorf(msg string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

func (l *Logger) Debug(msg string) {
	l.logger.Printf("[DEBUG] %s", msg)
}

func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}

func (l *Logger) Warn(msg string) {
	l.logger.Printf("[WARN] %s", msg)
}

func NewStdLogger(w io.Writer) *log.Logger {
	return log.New(w, "", log.LstdFlags)
}