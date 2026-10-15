package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	portsout "github.com/DevThiagoSouza/src/application/ports/out"
)


type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

type SimpleLogger struct {
	requestID string
	level     LogLevel
	logger    *log.Logger
}

func NewSimpleLogger(level LogLevel) *SimpleLogger {
	return &SimpleLogger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}


func (sl *SimpleLogger) formatMessage(level LogLevel, msg string, args ...interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)
	levelStr := levelToString(level)
	requestIDPart := ""
	if sl.requestID != "" {
		requestIDPart = fmt.Sprintf("[RequestID: %s] ", sl.requestID)
	}
	formattedMsg := msg
	if len(args) > 0 {
		formattedMsg = fmt.Sprintf(msg, args...)
	}

	return fmt.Sprintf("%s [%s] %s%s", timestamp, levelStr, requestIDPart, formattedMsg)
}

func (sl *SimpleLogger) Debug(ctx context.Context, msg string, args ...interface{}) {
	_ = ctx
	if sl.level <= DEBUG {
		formattedMsg := sl.formatMessage(DEBUG, msg, args...)
		sl.logger.Println(formattedMsg)
	}
}

func (sl *SimpleLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	_ = ctx
	if sl.level <= INFO {
		formattedMsg := sl.formatMessage(INFO, msg, args...)
		sl.logger.Println(formattedMsg)
	}
}

func (sl *SimpleLogger) Warning(ctx context.Context, msg string, args ...interface{}) {
	_ = ctx
	if sl.level <= WARNING {
		formattedMsg := sl.formatMessage(WARNING, msg, args...)
		sl.logger.Println(formattedMsg)
	}
}

func (sl *SimpleLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	_ = ctx
	if sl.level <= ERROR {
		formattedMsg := sl.formatMessage(ERROR, msg, args...)
		sl.logger.Println(formattedMsg)
	}
}

func (sl *SimpleLogger) WithRequestID(requestID string) portsout.Logger {
	return &SimpleLogger{
		requestID: requestID,
		level:     sl.level,
		logger:    sl.logger,
	}
}

