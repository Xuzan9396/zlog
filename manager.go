package zlog

import (
	"errors"
	"runtime/debug"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Manager 提供独立的日志实例，避免全局配置相互覆盖。
type Manager struct {
	cfgMu       sync.RWMutex
	cfg         Config
	level       zap.AtomicLevel
	registry    *loggerRegistry
	cleanupTask *CleanupTask
	cleanupOnce sync.Once
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

func getDefaultManager() *Manager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewManager()
	})
	return defaultManager
}

// NewManager 创建一个新的日志管理器，可选地应用配置选项。
func NewManager(options ...LogOption) *Manager {
	cfg := newDefaultConfig()
	applyOptions(&cfg, options...)

	mgr := &Manager{
		cfg:   cfg,
		level: zap.NewAtomicLevelAt(cfg.Level),
	}
	mgr.registry = newLoggerRegistry(&mgr.level, mgr.getConfig)

	// 延迟启动清理任务，避免初始化循环依赖
	if cfg.AutoCleanup {
		go func() {
			time.Sleep(100 * time.Millisecond)
			mgr.startCleanupTask()
		}()
	}

	return mgr
}

// SetLog 应用环境和选项到当前管理器。
func (m *Manager) SetLog(env Env, options ...LogOption) {
	m.cfgMu.Lock()
	cfg := m.cfg
	cfg.Env = env
	cfg.levelOverride = false
	applyOptions(&cfg, options...)
	m.cfg = cfg
	m.cfgMu.Unlock()

	m.level.SetLevel(cfg.Level)
	// 清除 logger 缓存，强制重新创建以应用新的输出配置
	m.registry.reset()

	// 重新启动清理任务（如果配置改变）
	m.startCleanupTask()
}

// UpdateRetention 更新日志保留及切割周期。
func (m *Manager) UpdateRetention(withMaxAge, withRotationTime int) {
	m.cfgMu.Lock()
	cfg := m.cfg
	cfg.WithMaxAge = withMaxAge
	cfg.WithRotationTime = withRotationTime
	m.cfg = cfg
	m.cfgMu.Unlock()
}

// SetLevel 显式设置日志等级，并覆盖环境默认值。
func (m *Manager) SetLevel(level zapcore.Level) {
	m.cfgMu.Lock()
	cfg := m.cfg
	cfg.Level = level
	cfg.levelOverride = true
	m.cfg = cfg
	m.cfgMu.Unlock()

	m.level.SetLevel(level)
	// 清除 logger 缓存，强制重新创建以应用新的输出配置
	m.registry.reset()
}

// Logger 返回指定名称的 SugaredLogger。
func (m *Manager) Logger(fileNameArr ...string) *zap.SugaredLogger {
	cfg := m.getConfig()
	name := cfg.DefaultLoggerName
	if len(fileNameArr) > 0 && fileNameArr[0] != "" {
		name = fileNameArr[0]
	}

	var skip uint8
	if len(fileNameArr) > 1 {
		skip = 1
	}

	return m.registry.getOrCreate(name, skip)
}

// F 是 Logger 的别名，方便与全局 API 对齐。
func (m *Manager) F(fileNameArr ...string) *zap.SugaredLogger {
	return m.Logger(fileNameArr...)
}

// Sync 刷新指定 logger 的缓冲区。
func (m *Manager) Sync(fileName string) error {
	defer func() {
		if err := recover(); err != nil {
			m.Logger().Errorf("错误panic:%s", string(debug.Stack()))
		}
	}()
	if fileName == "" {
		return errors.New("文件错误")
	}
	if logger, ok := m.registry.get(fileName); ok && logger != nil {
		return logger.Sync()
	}
	return nil
}

// SetDebugLevel 将等级调整为 Debug。
func (m *Manager) SetDebugLevel() {
	m.SetLevel(zapcore.DebugLevel)
}

// SetInfoLevel 将等级调整为 Info。
func (m *Manager) SetInfoLevel() {
	m.SetLevel(zapcore.InfoLevel)
}

// SetWarnLevel 将等级调整为 Warn。
func (m *Manager) SetWarnLevel() {
	m.SetLevel(zapcore.WarnLevel)
}

// SetErrorLevel 将等级调整为 Error。
func (m *Manager) SetErrorLevel() {
	m.SetLevel(zapcore.ErrorLevel)
}

// SetDPanicLevel 将等级调整为 DPanic。
func (m *Manager) SetDPanicLevel() {
	m.SetLevel(zapcore.DPanicLevel)
}

// SetPanicLevel 将等级调整为 Panic。
func (m *Manager) SetPanicLevel() {
	m.SetLevel(zapcore.PanicLevel)
}

// SetFatalLevel 将等级调整为 Fatal。
func (m *Manager) SetFatalLevel() {
	m.SetLevel(zapcore.FatalLevel)
}

// SetConsoleOnly 动态设置仅输出到终端模式。
func (m *Manager) SetConsoleOnly(consoleOnly bool) {
	m.cfgMu.Lock()
	cfg := m.cfg
	cfg.ConsoleOnly = consoleOnly
	m.cfg = cfg
	m.cfgMu.Unlock()

	// 清除 logger 缓存，强制重新创建以应用新的输出配置
	m.registry.reset()
}

// startCleanupTask 启动后台清理任务
func (m *Manager) startCleanupTask() {
	cfg := m.getConfig()

	// 如果禁用自动清理，停止现有任务
	if !cfg.AutoCleanup {
		m.StopCleanupTask()
		return
	}

	// 只初始化一次
	m.cleanupOnce.Do(func() {
		m.cleanupTask = newCleanupTask(cfg.CleanupInterval)
		m.cleanupTask.Start()
	})

	// 如果已经存在任务，检查是否需要更新
	if m.cleanupTask != nil {
		// 如果间隔改变，重启任务
		if m.cleanupTask.interval != cfg.CleanupInterval {
			m.StopCleanupTask()
			m.cleanupOnce = sync.Once{} // 重置 Once
			m.cleanupTask = newCleanupTask(cfg.CleanupInterval)
			m.cleanupTask.Start()
		}
	}
}

// CleanupLogs 手动触发一次日志清理
func (m *Manager) CleanupLogs() {
	clearLog()
}

// StopCleanupTask 停止后台清理任务
func (m *Manager) StopCleanupTask() {
	if m.cleanupTask != nil {
		m.cleanupTask.Stop()
	}
}

// IsCleanupRunning 返回清理任务是否正在运行
func (m *Manager) IsCleanupRunning() bool {
	if m.cleanupTask == nil {
		return false
	}
	return m.cleanupTask.IsRunning()
}

// getConfig 返回配置副本，供内部使用。
func (m *Manager) getConfig() Config {
	m.cfgMu.RLock()
	defer m.cfgMu.RUnlock()
	return cloneConfig(m.cfg)
}

// Public wrapper functions for backwards compatibility.

// SetLog 是兼容旧行为的全局入口。
func SetLog(env Env, options ...LogOption) {
	getDefaultManager().SetLog(env, options...)
}

// SetEnv 兼容旧行为，等价于 SetLog。
func SetEnv(env string) {
	getDefaultManager().SetLog(Env(env))
}

// SetConfig 兼容旧行为，仅调整保留和切割周期。
func SetConfig(withMaxAge, withRotationTime int) {
	getDefaultManager().UpdateRetention(withMaxAge, withRotationTime)
}

// getConfig 返回默认管理器的配置副本。
func getConfig() Config {
	return getDefaultManager().getConfig()
}

// SetZapOut 兼容旧行为，将标准库 log 输出重定向到默认管理器。
func SetZapOut(fileName string) error {
	return getDefaultManager().SetZapOut(fileName)
}

// CleanupLogs 手动触发全局日志清理。
func CleanupLogs() {
	getDefaultManager().CleanupLogs()
}

// StopCleanupTask 停止全局后台清理任务。
func StopCleanupTask() {
	getDefaultManager().StopCleanupTask()
}

// IsCleanupRunning 返回全局清理任务是否正在运行。
func IsCleanupRunning() bool {
	return getDefaultManager().IsCleanupRunning()
}
