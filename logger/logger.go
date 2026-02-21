package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var logLevelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = New(INFO)
}

func New(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) formatMessage(level LogLevel, msg string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] %s - %s", timestamp, logLevelNames[level], msg)
}

func (l *Logger) log(level LogLevel, msg string) {
	if level >= l.level {
		l.logger.Println(l.formatMessage(level, msg))
	}
}

func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(msg string) {
	l.log(INFO, msg)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(msg string) {
	l.log(WARN, msg)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(msg string) {
	l.log(ERROR, msg)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(msg string) {
	l.log(FATAL, msg)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func Debug(msg string) {
	defaultLogger.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(msg string) {
	defaultLogger.Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func SetLevel(level LogLevel) {
	defaultLogger.level = level
}
