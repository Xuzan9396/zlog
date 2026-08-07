package zlog

import (
	"sync/atomic"
	"testing"
	"time"
)

// prepareWatchTest 为每个监听测试创建独立文件日志，避免受到其他测试 ConsoleOnly 和日志目录状态影响。
// 例: 先写一条初始化错误确保当天文件存在，再等待 tail 协程就绪后才发送待断言日志。
func prepareWatchTest(t *testing.T) {
	t.Helper()
	if allcancel != nil {
		allcancel()
		allcancel = nil
	}
	waitTailStopped(t)

	orig := getConfig()
	// 先注册 TempDir 自带的清理，再注册自定义清理。testing 按后进先出执行清理，
	// 因此 tail 会先停止，避免临时目录删除后监听协程仍尝试重开文件。
	testLogDir := t.TempDir()
	t.Cleanup(func() {
		if allcancel != nil {
			allcancel()
			allcancel = nil
		}
		waitTailStopped(t)
		SetLog(orig.Env,
			WithMaxAge(orig.WithMaxAge),
			WithRotationTime(orig.WithRotationTime),
			WithDate(orig.formDate),
			WithConsoleOnly(orig.ConsoleOnly),
			WithAutoCleanup(orig.AutoCleanup),
			WithCleanupInterval(orig.CleanupInterval),
			WithLogDir(orig.LogDir),
			WithDefaultName(orig.DefaultLoggerName),
			WithErrorName(orig.ErrorLoggerName),
		)
	})

	SetLog(ENV_INFO, WithConsoleOnly(false), WithLogDir(testLogDir))
	F().Error("初始化错误日志监听文件")
	if err := Sync(getConfig().DefaultLoggerName); err != nil {
		t.Fatalf("初始化错误日志文件同步失败: %v", err)
	}
}

// waitTailStopped 等待上一条 tail 协程完成退出，不能直接把计数器清零，
// 否则旧协程稍后执行递减时会导致 uint32 下溢并污染后续测试。
func waitTailStopped(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for atomic.LoadUint32(&curTailNum) != 0 {
		if time.Now().After(deadline) {
			t.Fatal("等待上一条错误日志 tail 协程退出超时")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func waitTailReady(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for atomic.LoadUint32(&curTailNum) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("等待错误日志 tail 协程就绪超时")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWatchErr(t *testing.T) {
	prepareWatchTest(t)
	ch, err := WatchErr()
	if err != nil {
		t.Fatal(err)
	}
	waitTailReady(t)

	go func() {
		for i := 0; i < 3; i++ {
			F().Errorf("watch err test: %d", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case msg := <-ch:
		if msg == "" {
			t.Fatal("received empty message")
		}
		t.Log(msg)
	case <-time.After(5 * time.Second):
		t.Fatal("watch err timeout")
	}
}

func TestWatchErrCallback(t *testing.T) {
	prepareWatchTest(t)
	msgs := make(chan string, 1)

	if err := WatchErrCallback(func(msg string) {
		select {
		case msgs <- msg:
		default:
		}
	}); err != nil {
		t.Fatal(err)
	}
	waitTailReady(t)

	go func() {
		for i := 0; i < 3; i++ {
			F().Errorf("watch err callback test: %d", i)
			time.Sleep(200 * time.Millisecond)
		}
	}()

	select {
	case msg := <-msgs:
		if msg == "" {
			t.Fatal("callback received empty message")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("watch err callback timeout")
	}
}
