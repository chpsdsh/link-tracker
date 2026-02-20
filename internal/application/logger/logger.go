package logger

import (
	"io"
	"log/slog"
)

type OutputFormat int

const (
	OutputFormatJson OutputFormat = iota
	OutputFormatText
)

func NewLogger(writer io.Writer, format OutputFormat, level slog.Level) *slog.Logger {
	var log slog.Handler
	switch format {
	case OutputFormatJson:
		log = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: level})
	case OutputFormatText:
		log = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})
	}
	return slog.New(log)
}
