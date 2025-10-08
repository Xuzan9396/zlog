package zlog

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type configProvider func() Config

// loggerRegistry 负责缓存 SugaredLogger，避免重复创建 zap Core。
type loggerRegistry struct {
	mu          sync.RWMutex
	loggers     map[string]*zap.SugaredLogger
	errorWriter zapcore.WriteSyncer
	errorOnce   sync.Once
	level       *zap.AtomicLevel
	cfgFn       configProvider
}

// newLoggerRegistry 构造一个空的日志注册表。
func newLoggerRegistry(level *zap.AtomicLevel, cfgFn configProvider) *loggerRegistry {
	return &loggerRegistry{
		loggers: make(map[string]*zap.SugaredLogger),
		level:   level,
		cfgFn:   cfgFn,
	}
}

// get 返回已存在的 logger，未命中时返回 false。
func (r *loggerRegistry) get(name string) (*zap.SugaredLogger, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	logger, ok := r.loggers[name]
	return logger, ok
}

// getOrCreate 获取或懒加载指定名称的 logger。
func (r *loggerRegistry) getOrCreate(name string, skipCaller uint8) *zap.SugaredLogger {
	if logger, ok := r.get(name); ok {
		return logger
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if logger, ok := r.loggers[name]; ok {
		return logger
	}

	logger := r.buildLogger(name, skipCaller)
	r.loggers[name] = logger
	return logger
}

// buildLogger 构建底层 zap core，并根据环境级别配置 writer。
func (r *loggerRegistry) buildLogger(name string, skipCaller uint8) *zap.SugaredLogger {
	cfg := r.cfgFn()

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "line",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     newTimeEncoder(r.cfgFn),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	infoWriter, _ := newInfoWriter(cfg, logFilePath("%s_info.log", name))

	errorWriter := r.ensureErrorWriter(cfg)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), r.level),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), zapcore.ErrorLevel),
	)

	caller := []zap.Option{zap.AddCaller()}
	if skipCaller > 0 {
		caller = append(caller, zap.AddCallerSkip(int(skipCaller)))
	}

	return zap.New(core, caller...).Sugar()
}

// ensureErrorWriter 构建共享的 error writer，保证只初始化一次。
func (r *loggerRegistry) ensureErrorWriter(cfg Config) zapcore.WriteSyncer {
	r.errorOnce.Do(func() {
		writer, err := newErrorWriter(cfg, logFilePath("%s.log", cfg.ErrorLoggerName))
		if err != nil || writer == nil {
			fmt.Fprintf(os.Stderr, "zlog: failed to create error writer: %v\n", err)
			r.errorWriter = zapcore.AddSync(os.Stderr)
		} else {
			r.errorWriter = writer
		}
	})
	return r.errorWriter
}

func (r *loggerRegistry) resetErrorWriter() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errorWriter = nil
	r.errorOnce = sync.Once{}
}
