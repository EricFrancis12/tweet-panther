package main

import (
	"fmt"
	"time"
)

type Logger struct{}

func newLogger() *Logger {
	return &Logger{}
}

func (l Logger) Log(level LogLevel, message string) {
	timestamp := time.Now().Format(time.UnixDate)
	if level == LogLevelInfo {
		fmt.Printf("[%s] %s", timestamp, message)
		return
	}
	fmt.Printf("[%s] [%s] %s", timestamp, level, message)
}

func (l Logger) Info(message string) {
	l.Log(LogLevelInfo, message)
}

func (l Logger) Infof(format string, a ...any) {
	l.Log(LogLevelInfo, fmt.Sprintf(format, a...))
}

func (l Logger) Infoln(message string) {
	l.Log(LogLevelInfo, fmt.Sprintf("%s\n", message))
}

func (l Logger) Error(message string) {
	l.Log(LogLevelError, message)
}

func (l Logger) Errorf(format string, a ...any) {
	l.Log(LogLevelError, fmt.Sprintf(format, a...))
}

func (l Logger) Errorln(message string) {
	l.Log(LogLevelError, fmt.Sprintf("%s\n", message))
}

func (l Logger) LogErr(err error) {
	if err != nil {
		l.Errorln(err.Error())
	}
}
