package zlog

import (
	"log"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestSetLogAppliesOptions(t *testing.T) {
	SetLog(ENV_DEBUG, WithMaxAge(12), WithRotationTime(6), WithDate(DATE_MSEC))
	t.Cleanup(func() {
		SetLog(ENV_PRO, WithMaxAge(10*24), WithRotationTime(24), WithDate(DATE_SEC))
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

	SetLog(ENV_PRO, WithMaxAge(12), WithRotationTime(6), WithDate(DATE_MSEC))
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
