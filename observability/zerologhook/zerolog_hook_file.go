package zerologhook

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

func NewRotatingWriter(filename string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,  // in MB
		MaxBackups: maxBackups, //  file backup
		MaxAge:     maxAgeDays, // in days
		Compress:   compress,
	}
}
