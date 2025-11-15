package loghook

import (
	"fmt"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type rotatingWriter struct {
	errorLogFile *lumberjack.Logger
	logFile      *lumberjack.Logger
}

func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), `"level":"ERROR"`) {
		return w.errorLogFile.Write(p)
	}

	return w.logFile.Write(p)
}

func (w *rotatingWriter) Close() {
	w.errorLogFile.Close()
	w.logFile.Close()
}

func NewRotatingWriter(filename string, maxSizeMB, maxBackups, maxAgeDays int, compress bool) *rotatingWriter {
	return &rotatingWriter{
		logFile: &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSizeMB,  // in MB
			MaxBackups: maxBackups, //  file backup
			MaxAge:     maxAgeDays, // in days
			Compress:   compress,
		},
		errorLogFile: &lumberjack.Logger{
			Filename:   fmt.Sprintf("error_%s", filename),
			MaxSize:    maxSizeMB,  // in MB
			MaxBackups: maxBackups, //  file backup
			MaxAge:     maxAgeDays, // in days
			Compress:   compress,
		},
	}
}
