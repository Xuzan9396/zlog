//go:build !windows
// +build !windows

package zlog

import (
	"log"
	"os"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
)

// newInfoWriter 创建 info 日志的滚动写入器，支持控制台双写。
func newInfoWriter(cfg Config, fileName string) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(time.Duration(cfg.WithMaxAge)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(cfg.WithRotationTime)*time.Hour),
	)
	if cfg.Env == ENV_DEBUG {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
	return zapcore.AddSync(fileWriter), err
}

// newErrorWriter 创建 error 日志专用的滚动写入器。
func newErrorWriter(cfg Config, fileName string) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(time.Duration(cfg.WithMaxAge)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(cfg.WithRotationTime)*time.Hour),
	)
	return zapcore.AddSync(fileWriter), err
}

// SetZapOut 将标准库 log 输出重定向到滚动日志。
func (m *Manager) SetZapOut(fileName string) error {
	cfg := m.getConfig()

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	fileWriter, err := rotatelogs.New(
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log",
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithRotationCount(7),
		rotatelogs.WithRotationSize(1024*1024*10),
	)
	var w zapcore.WriteSyncer
	if cfg.Env == ENV_DEBUG {
		w, err = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
		if err != nil {
			return err
		}
	} else {
		w = zapcore.AddSync(fileWriter)
	}
	log.SetOutput(w)
	return nil
}
