package zlog

import "go.uber.org/zap/zapcore"

// Env 表示日志运行的环境或预设等级。
type Env string

// EnvDate 定义日志写入的时间格式模板。
type EnvDate string

const (
	LOG_PRO    = "pro"
	LOG_DEBUG  = "debug"
	LOG_INFO   = "info"
	LOG_WARN   = "warn"
	LOG_ERROR  = "error"
	LOG_DPANIC = "dpanic"
	LOG_PANIC  = "panic"
	LOG_FATAL  = "fatal"

	ENV_PRO    Env = Env(LOG_PRO)
	ENV_DEBUG  Env = Env(LOG_DEBUG)
	ENV_INFO   Env = Env(LOG_INFO)
	ENV_WARN   Env = Env(LOG_WARN)
	ENV_ERROR  Env = Env(LOG_ERROR)
	ENV_DPANIC Env = Env(LOG_DPANIC)
	ENV_PANIC  Env = Env(LOG_PANIC)
	ENV_FATAL  Env = Env(LOG_FATAL)

	DATE_SEC  EnvDate = "2006-01-02 15:04:05"
	DATE_MSEC EnvDate = "2006-01-02 15:04:05.000"
)

// Config 聚合日志系统运行所需的全部配置。
type Config struct {
	WithMaxAge       int
	WithRotationTime int
	Env              Env
	Level            zapcore.Level
	formDate         EnvDate
	levelOverride    bool
}

// LogOption 通过函数式选项修改配置。
type LogOption func(*Config)

var envLevelMap = map[Env]zapcore.Level{
	ENV_PRO:    zapcore.InfoLevel,
	ENV_DEBUG:  zapcore.DebugLevel,
	ENV_INFO:   zapcore.InfoLevel,
	ENV_WARN:   zapcore.WarnLevel,
	ENV_ERROR:  zapcore.ErrorLevel,
	ENV_DPANIC: zapcore.DPanicLevel,
	ENV_PANIC:  zapcore.PanicLevel,
	ENV_FATAL:  zapcore.FatalLevel,
}

// newDefaultConfig 返回默认配置。
func newDefaultConfig() Config {
	return Config{
		WithMaxAge:       10 * 24,
		WithRotationTime: 24,
		Env:              ENV_PRO,
		Level:            resolveLevel(ENV_PRO),
		formDate:         DATE_SEC,
	}
}

// cloneConfig 生成配置副本，避免外部修改内部状态。
func cloneConfig(cfg Config) Config {
	return cfg
}

// WithMaxAge 设置日志归档保留时长（小时）。
func WithMaxAge(withMaxAge int) LogOption {
	return func(cfg *Config) {
		cfg.WithMaxAge = withMaxAge
	}
}

// WithRotationTime 设置日志切割周期（小时）。
func WithRotationTime(withRotationTime int) LogOption {
	return func(cfg *Config) {
		cfg.WithRotationTime = withRotationTime
	}
}

// WithDate 设置日志时间格式，如秒或毫秒模板。
func WithDate(date EnvDate) LogOption {
	return func(cfg *Config) {
		cfg.formDate = date
	}
}

// WithLevel 允许直接指定 zapcore.Level，用于高度自定义场景。
func WithLevel(level zapcore.Level) LogOption {
	return func(cfg *Config) {
		cfg.Level = level
		cfg.levelOverride = true
	}
}

// applyOptions 应用可选参数，并在未覆盖等级时同步环境默认等级。
func applyOptions(cfg *Config, options ...LogOption) {
	for _, option := range options {
		option(cfg)
	}
	if !cfg.levelOverride {
		cfg.Level = resolveLevel(cfg.Env)
	}
}

// resolveLevel 将 Env 映射到对应 zap 等级，默认返回 Info。
func resolveLevel(env Env) zapcore.Level {
	if level, ok := envLevelMap[env]; ok {
		return level
	}
	return zapcore.InfoLevel
}
