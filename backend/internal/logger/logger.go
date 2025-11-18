package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel string

const (
	LevelDebug   LogLevel = "DEBUG"
	LevelInfo    LogLevel = "INFO"
	LevelWarning LogLevel = "WARNING"
	LevelError   LogLevel = "ERROR"
	LevelFatal   LogLevel = "FATAL"
)

type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Error     string                 `json:"error,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	Path      string                 `json:"path,omitempty"`
	Method    string                 `json:"method,omitempty"`
	IP        string                 `json:"ip,omitempty"`
}

var (
	logFile   *os.File
	logDir    = "/var/log/youtube-market"
	errorFile *os.File
)

// Init инициализирует систему логирования
func Init() error {
	// Создаем директорию для логов
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Если не удалось создать в /var/log, используем локальную директорию
		logDir = "./logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Открываем файл для общих логов
	logPath := filepath.Join(logDir, "app.log")
	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Открываем файл для ошибок
	errorPath := filepath.Join(logDir, "errors.log")
	errorFile, err = os.OpenFile(errorPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open error log file: %w", err)
	}

	return nil
}

// Close закрывает файлы логов
func Close() {
	if logFile != nil {
		logFile.Close()
	}
	if errorFile != nil {
		errorFile.Close()
	}
}

// Log записывает лог-запись
func Log(level LogLevel, message string, err error, context map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Context:   context,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// Сериализуем в JSON
	jsonData, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		log.Printf("Failed to marshal log entry: %v", jsonErr)
		return
	}

	// Записываем в общий лог
	if logFile != nil {
		fmt.Fprintln(logFile, string(jsonData))
	}

	// Записываем ошибки в отдельный файл
	if level == LevelError || level == LevelFatal {
		if errorFile != nil {
			fmt.Fprintln(errorFile, string(jsonData))
		}
		// Также выводим в stderr
		log.Printf("[%s] %s: %v", level, message, err)
	} else {
		// Обычные логи в stdout
		log.Printf("[%s] %s", level, message)
	}
}

// Debug логирует отладочное сообщение
func Debug(message string, context map[string]interface{}) {
	Log(LevelDebug, message, nil, context)
}

// Info логирует информационное сообщение
func Info(message string, context map[string]interface{}) {
	Log(LevelInfo, message, nil, context)
}

// Warning логирует предупреждение
func Warning(message string, context map[string]interface{}) {
	Log(LevelWarning, message, nil, context)
}

// Error логирует ошибку
func Error(message string, err error, context map[string]interface{}) {
	Log(LevelError, message, err, context)
}

// Fatal логирует критическую ошибку и завершает программу
func Fatal(message string, err error, context map[string]interface{}) {
	Log(LevelFatal, message, err, context)
	os.Exit(1)
}

// ErrorWithContext логирует ошибку с контекстом запроса
func ErrorWithContext(message string, err error, userID, path, method, ip string, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["user_id"] = userID
	context["path"] = path
	context["method"] = method
	context["ip"] = ip

	Log(LevelError, message, err, context)
}

// GetLogDir возвращает директорию с логами
func GetLogDir() string {
	return logDir
}

