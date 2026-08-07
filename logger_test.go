package zlog

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestSetLogAppliesOptions(t *testing.T) {
	SetLog(ENV_DEBUG, WithMaxAge(12), WithRotationTime(6), WithDate(DATE_MSEC), WithConsoleOnly(false))
	t.Cleanup(func() {
		SetLog(ENV_PRO, WithMaxAge(10*24), WithRotationTime(24), WithDate(DATE_SEC), WithConsoleOnly(false))
	})

	cfg := getConfig()
	if cfg.Env != ENV_DEBUG {
		t.Fatalf("expected env %s, got %s", ENV_DEBUG, cfg.Env)
	}
	if cfg.WithMaxAge != 12 {
		t.Fatalf("expected max age 12, got %d", cfg.WithMaxAge)
	}
	if cfg.WithRotationTime != 6 {
		t.Fatalf("expected rotation time 6, got %d", cfg.WithRotationTime)
	}
	if cfg.formDate != DATE_MSEC {
		t.Fatalf("expected DATE_MSEC, got %s", cfg.formDate)
	}
	if cfg.Level != zapcore.DebugLevel {
		t.Fatalf("expected debug level, got %s", cfg.Level)
	}

	SetLog(ENV_PRO, WithMaxAge(12), WithRotationTime(6), WithDate(DATE_MSEC), WithConsoleOnly(false))
	logger := F("setlog")
	logger.Debug("test message")
	if err := Sync("setlog"); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
}

func TestSyncUnknownLogger(t *testing.T) {
	if err := Sync("unknown"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSetLogDirAndWithLogDir(t *testing.T) {
	orig := getConfig()
	t.Cleanup(func() {
		SetLog(
			orig.Env,
			WithMaxAge(orig.WithMaxAge),
			WithRotationTime(orig.WithRotationTime),
			WithDate(orig.formDate),
			WithLogDir(orig.LogDir),
		)
	})

	tempDir := t.TempDir()
	mgr := NewManager(WithLogDir(tempDir), WithAutoCleanup(false))
	if got := mgr.getConfig().LogDir; got != tempDir {
		t.Fatalf("expected manager log dir %s, got %s", tempDir, got)
	}
	if got := logDir(); got != tempDir {
		t.Fatalf("expected global log dir %s, got %s", tempDir, got)
	}

	nextDir := filepath.Join(t.TempDir(), "logs")
	if err := mgr.SetLogDir(nextDir); err != nil {
		t.Fatalf("SetLogDir failed: %v", err)
	}
	if got := mgr.getConfig().LogDir; got != nextDir {
		t.Fatalf("expected updated manager log dir %s, got %s", nextDir, got)
	}
	if got := logDir(); got != nextDir {
		t.Fatalf("expected updated global log dir %s, got %s", nextDir, got)
	}

	if err := SetLogDir(""); err == nil {
		t.Fatal("expected error for empty log dir")
	}
}

// TestNewManagerDefersLogDirCreation 验证只创建 Manager 不会污染工作目录或提前创建目标目录。
// 例: 应用先完成配置、稍后首次写日志时，目录才由 Writer 创建。
func TestNewManagerDefersLogDirCreation(t *testing.T) {
	origDir := logDir()
	t.Cleanup(func() { setLogDir(origDir) })
	targetDir := filepath.Join(t.TempDir(), "deferred-logs")
	mgr := NewManager(WithLogDir(targetDir), WithAutoCleanup(false))
	t.Cleanup(mgr.StopCleanupTask)
	// SetLog 只更新运行配置；应用尚未真正写日志时也不能产生目录副作用。
	mgr.SetLog(ENV_PRO, WithMaxAge(2*24))

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Fatalf("NewManager/SetLog 提前创建了日志目录，路径=%s，错误=%v", targetDir, err)
	}

	logger := mgr.Logger("deferred")
	logger.Info("create on first write")
	if err := mgr.Sync("deferred"); err != nil {
		t.Fatalf("首次写入后同步日志失败: %v", err)
	}
	if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
		t.Fatalf("首次写入没有创建日志目录，路径=%s，错误=%v", targetDir, err)
	}
}

func TestSetZapOutWrites(t *testing.T) {
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "sys_log.log")
	if err := SetZapOut(target); err != nil {
		t.Fatalf("SetZapOut failed: %v", err)
	}

	log.Println("testing std logger output")
	time.Sleep(200 * time.Millisecond)
}

func TestResolveLevelMapping(t *testing.T) {
	orig := getConfig()
	t.Cleanup(func() {
		SetLog(
			orig.Env,
			WithMaxAge(orig.WithMaxAge),
			WithRotationTime(orig.WithRotationTime),
			WithDate(orig.formDate),
			WithLevel(orig.Level),
		)
	})

	cases := map[Env]zapcore.Level{
		ENV_PRO:    zapcore.InfoLevel,
		ENV_DEBUG:  zapcore.DebugLevel,
		ENV_INFO:   zapcore.InfoLevel,
		ENV_WARN:   zapcore.WarnLevel,
		ENV_ERROR:  zapcore.ErrorLevel,
		ENV_DPANIC: zapcore.DPanicLevel,
		ENV_PANIC:  zapcore.PanicLevel,
		ENV_FATAL:  zapcore.FatalLevel,
	}

	for env, want := range cases {
		SetLog(env)
		if got := getConfig().Level; got != want {
			t.Fatalf("env %s expected level %s, got %s", env, want, got)
		}
	}
}

func TestManagerIsolation(t *testing.T) {
	mgrA := NewManager()
	mgrB := NewManager()

	mgrA.SetLog(ENV_DEBUG)
	if mgrA.getConfig().Env != ENV_DEBUG {
		t.Fatalf("manager A env not updated")
	}
	if mgrB.getConfig().Env != ENV_PRO {
		t.Fatalf("manager B env should remain default, got %s", mgrB.getConfig().Env)
	}

	loggerA := mgrA.Logger("isolation")
	loggerB := mgrB.Logger("isolation")
	if loggerA == loggerB {
		t.Fatal("expected different logger instances for different managers")
	}

	mgrA.SetLevel(zapcore.ErrorLevel)
	if lvl := mgrA.getConfig().Level; lvl != zapcore.ErrorLevel {
		t.Fatalf("manager A level not updated, got %s", lvl)
	}
	if lvl := mgrB.getConfig().Level; lvl != resolveLevel(mgrB.getConfig().Env) {
		t.Fatalf("manager B level changed unexpectedly: %s", lvl)
	}
}

func TestWithDefaultNameOption(t *testing.T) {
	mgr := NewManager(WithDefaultName("log"), WithErrorName("log_err"))
	cfg := mgr.getConfig()
	if cfg.DefaultLoggerName != "log" {
		t.Fatalf("expected default name log, got %s", cfg.DefaultLoggerName)
	}
	if cfg.ErrorLoggerName != "log_err" {
		t.Fatalf("expected error name log_err, got %s", cfg.ErrorLoggerName)
	}

	logger := mgr.Logger()
	if logger == nil {
		t.Fatal("expected logger instance")
	}
}
