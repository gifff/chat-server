package logger

import (
	"io"

	"github.com/hashicorp/logutils"
)

func NewLevelFilter(level string, w io.Writer) io.Writer {
	var minLogLevel logutils.LogLevel

	switch level {
	case "DEBUG", "INFO", "WARN", "ERROR":
		minLogLevel = logutils.LogLevel(level)
	default:
		minLogLevel = logutils.LogLevel("INFO")
	}

	return &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: minLogLevel,
		Writer:   w,
	}
}
