package loghook

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type rotatingWriter struct {
	errorLogFile *lumberjack.Logger
	logFile      *lumberjack.Logger

	serviceNamespace  string
	serviceName       string
	serviceInstanceID string
	serviceVersion    string
	env               string
}

func (w *rotatingWriter) Write(p []byte) (n int, err error) {
	if isErrorLogLine(p) {
		return w.errorLogFile.Write(p)
	}

	return w.logFile.Write(p)
}

func isErrorLogLine(p []byte) bool {
	return bytes.Contains(p, []byte(`"level":"error"`)) ||
		bytes.Contains(p, []byte(`"level":"fatal"`)) ||
		bytes.Contains(p, []byte(`"level":"panic"`)) ||
		bytes.Contains(p, []byte(`"level":"ERROR"`))
}

func (w *rotatingWriter) Close() {
	w.errorLogFile.Close()
	w.logFile.Close()
	slog.Info("close rotating log writer successfully",
		slog.String("service", w.serviceName),
		slog.String("namespace", w.serviceNamespace),
		slog.String("env", w.env),
	)
}

func NewRotatingWriter(filename string, maxSizeMB, maxBackups, maxAgeDays int, compress bool,
	serviceNamespace, serviceName, serviceInstanceID, serviceVersion, env string,
) *rotatingWriter {
	onlyFileName := strings.Split(filename, ".")[0]
	return &rotatingWriter{
		serviceNamespace:  serviceNamespace,
		serviceName:       serviceName,
		serviceInstanceID: serviceInstanceID,
		serviceVersion:    serviceVersion,
		env:               env,
		logFile: &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    maxSizeMB,  // in MB
			MaxBackups: maxBackups, //  file backup
			MaxAge:     maxAgeDays, // in days
			Compress:   compress,
		},
		errorLogFile: &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s-error.log", onlyFileName),
			MaxSize:    maxSizeMB,  // in MB
			MaxBackups: maxBackups, //  file backup
			MaxAge:     maxAgeDays, // in days
			Compress:   compress,
		},
	}
}
