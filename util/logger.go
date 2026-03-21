package util

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
)

type Level int

const (
	Debug Level = iota
	Info
	Warn
	Error
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	name    string
	level   Level
	enabled bool
	logger  *log.Logger
	loggers *slog.Logger
}

func NewLogger(name string, out io.Writer, level Level, enabled bool) *Logger {
	return &Logger{
		name:    name,
		level:   level,
		enabled: enabled,
		logger:  log.New(out, "", log.LstdFlags|log.Lmicroseconds),
		loggers: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

func (l *Logger) SetEnabled(enabled bool) {
	l.enabled = enabled
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) logf(level Level, format string, args ...any) {
	if !l.enabled {
		return
	}
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] [%s] %s", level.String(), l.name, msg)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(Debug, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(Info, format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logf(Warn, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(Error, format, args...)
}
