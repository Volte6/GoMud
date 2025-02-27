package mudlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/volte6/gomud/internal/events"
)

var (
	slogInstance *slog.Logger
)

func SetupLogger(inGameLogger events.EventLogger, logLevel string, logPath string, colorLogs bool) {

	SetLogLevel(strings.ToUpper(logLevel))

	// No filepath? Write to Stderr.
	if logPath == `` {
		slogInstance = slog.New(
			getLogHandler(os.Stderr, inGameLogger, colorLogs),
		)
		return
	}

	// Setup file logging

	fileInfo, err := os.Stat(logPath)
	if err == nil {
		if fileInfo.IsDir() {
			panic(fmt.Errorf("log file path is a directory: %s", logPath))
		}

	} else if os.IsNotExist(err) {
		// File does not exist; check if the directory exists
		dir := filepath.Dir(logPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			panic(fmt.Errorf("directory for log file does not exist: %s", dir))
		}
	} else {
		// Some other error
		panic(fmt.Errorf("error accessing log file path: %v", err))
	}

	lj := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,  // Maximum size in megabytes before rotation
		MaxBackups: 10,   // Maximum number of old log files to retain
		Compress:   true, // Compress rotated files
	}

	// Open or create the log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Errorf("failed to open log file: %v", err))
	}
	defer file.Close()

	slogInstance = slog.New(
		getLogHandler(lj, inGameLogger, colorLogs),
	)

}

func Debug(msg string, args ...any) {
	slogInstance.Log(context.Background(), slog.LevelDebug, msg, args...)
}

func Info(msg string, args ...any) {
	slogInstance.Log(context.Background(), slog.LevelInfo, msg, args...)
}

func Warn(msg string, args ...any) {
	slogInstance.Log(context.Background(), slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	slogInstance.Log(context.Background(), slog.LevelError, msg, args...)
}
