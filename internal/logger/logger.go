package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logger      *slog.Logger
	mu          sync.Mutex
	logFile     *os.File
	currentDate string
	logDir      string
)

func Init(dir string) error {
	mu.Lock()
	defer mu.Unlock()

	logDir = dir

	var writer io.Writer
	writer = os.Stdout
	logFile = nil

	if dir != "" {
		if err := os.MkdirAll(logDir, 0755); err == nil {
			filename := filepath.Join(logDir, "app-"+time.Now().Format("2006-01-02")+".log")
			file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				logFile = file
				writer = io.MultiWriter(os.Stdout, logFile)
			}
		}
	}

	logger = slog.New(slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	if logFile != nil {
		logger.Info("logger initialized", "log_dir", logDir, "has_log_file", true)
	} else {
		logger.Info("logger initialized", "has_log_file", false)
	}
	return nil
}

func rotateLogFile() error {
	if logDir == "" || logFile == nil {
		return nil
	}

	today := time.Now().Format("2006-01-02")
	if today == currentDate {
		return nil
	}

	logFile.Close()

	filename := filepath.Join(logDir, "app-"+today+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logFile = file
	currentDate = today
	return nil
}

func checkAndRotate() {
	if logDir == "" || logFile == nil {
		return
	}

	today := time.Now().Format("2006-01-02")
	if today != currentDate {
		mu.Lock()
		defer mu.Unlock()
		rotateLogFile()
	}
}

func Info(msg string, args ...any) {
	checkAndRotate()
	logger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	checkAndRotate()
	logger.Warn(msg, args...)
}

func Debug(msg string, args ...any) {
	checkAndRotate()
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	checkAndRotate()
	logger.Error(msg, args...)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}
