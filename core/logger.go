package core

import (
	"io"
	"log"
	"os"
	"sync"
)

// LogLevel определяет уровень логирования
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger - структура логгера с уровнем и мьютексом для потокобезопасности
type Logger struct {
	logger *log.Logger
	level  LogLevel
	mutex  sync.Mutex
	out    io.Writer
}

// NewLogger создает новый Logger с указанным уровнем и выводом
func NewLogger(level LogLevel, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}
	return &Logger{
		logger: log.New(output, "[GANYMEDE] ", log.LstdFlags|log.Lmsgprefix),
		level:  level,
		out:    output,
	}
}

// SetLevel устанавливает уровень логирования
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// Debug пишет сообщение уровня DEBUG
func (l *Logger) Debug(msg string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.level <= DEBUG {
		l.logger.SetPrefix("[DEBUG] ")
		l.logger.Println(msg)
	}
}

// Info пишет сообщение уровня INFO
func (l *Logger) Info(msg string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.level <= INFO {
		l.logger.SetPrefix("[INFO] ")
		l.logger.Println(msg)
	}
}

// Warn пишет сообщение уровня WARN
func (l *Logger) Warn(msg string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.level <= WARN {
		l.logger.SetPrefix("[WARN] ")
		l.logger.Println(msg)
	}
}

// Error пишет сообщение уровня ERROR
func (l *Logger) Error(msg string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.level <= ERROR {
		l.logger.SetPrefix("[ERROR] ")
		l.logger.Println(msg)
	}
}

// Global логгер по умолчанию
var DefaultLogger = NewLogger(INFO, nil) // или
var logger = NewLogger(INFO, nil)
