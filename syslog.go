package zlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type CustomLogger struct {
	*log.Logger
}

func NewCustomLogger() *CustomLogger {
	return &CustomLogger{log.New(os.Stderr, "", log.LstdFlags)}
}

func (l *CustomLogger) Info(format string, v ...interface{}) {
	l.Output(3, fmt.Sprintf("[INFO] %s:%d %s(): %s", getCallerInfo(), getLineNumber(), getFunctionName(), fmt.Sprintf(format, v...)))
}

func (l *CustomLogger) Error(format string, v ...interface{}) {
	l.Output(3, fmt.Sprintf("[ERROR] %s:%d %s(): %s", getCallerInfo(), getLineNumber(), getFunctionName(), fmt.Sprintf(format, v...)))
}

func getCallerInfo() string {
	_, file, _, ok := runtime.Caller(3)
	if !ok {
		file = "???"
	} else {
		file = filepath.Base(file)
	}
	return file
}

func getFunctionName() string {
	pc, _, _, ok := runtime.Caller(3)
	if !ok {
		return "???"
	}
	return runtime.FuncForPC(pc).Name()
}

func getLineNumber() int {
	_, _, line, ok := runtime.Caller(3)
	if !ok {
		return 0
	}
	return line
}
