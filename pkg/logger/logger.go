package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Logger is a thin wrapper around zerolog.Logger.
type Logger struct {
	l zerolog.Logger
}

// NewLogger creates a configured zerolog logger. env should be "production" or "development".
func NewLogger(level, env string, out io.Writer) *Logger {
	if out == nil {
		out = os.Stdout
	}

	var w io.Writer = out
	if strings.ToLower(env) != "production" {
		// use human friendly console writer for non-production
		w = zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
	}

	zl := zerolog.New(w).With().Timestamp().Logger()

	// set level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &Logger{l: zl}
}

func (lg *Logger) Debug(msg string, fields map[string]interface{}) {
	e := lg.l.Debug()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	e.Msg(msg)
}

func (lg *Logger) Info(msg string, fields map[string]interface{}) {
	e := lg.l.Info()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	e.Msg(msg)
}

func (lg *Logger) Warn(msg string, fields map[string]interface{}) {
	e := lg.l.Warn()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	e.Msg(msg)
}

func (lg *Logger) Error(msg string, fields map[string]interface{}) {
	e := lg.l.Error()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	e.Msg(msg)
}

// Fatal logs and exits.
func (lg *Logger) Fatal(msg string, fields map[string]interface{}) {
	e := lg.l.Fatal()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	e.Msg(msg)
}

// WithFields returns a new zerolog event with the provided fields.
func (lg *Logger) WithFields(fields map[string]interface{}) *zerolog.Event {
	e := lg.l.Info()
	for k, v := range fields {
		e = e.Interface(k, v)
	}
	return e
}

// GetZerologLogger returns the underlying zerolog.Logger instance.
// This is useful when you need to pass the logger to a function that expects a zerolog.Logger.
func (lg *Logger) GetZerologLogger() *zerolog.Logger {
	return &lg.l
}
