package zlog

import (
	"testing"
	"time"
)

func Test_getLogDate(t *testing.T) {
	logFileName := "new_logs_info2024-03-16.log"

	_, ts, err := getLogDate(logFileName)
	if err != nil {
		t.Error(err)
		return
	}
	var LOC, _ = time.LoadLocation("Local")
	now := time.Now() // 确保now使用和ts相同的时区

	// 获取当前时间的年、月、日部分，并将小时、分钟、秒和纳秒设置为0，以便进行日期比较
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, LOC)

	//// 计算两个日期之间的差异天数，忽略时间部分
	//daysDiff := midnight.Sub(*ts).Hours() / 24
	//t.Log(ts.Unix(), midnight.Unix(), daysDiff)
	//
	//t.Logf("%s %v %v", logFileName, ts.Format("2006-01-02"), daysDiff)

	// 判断ts与当前时间的差值是否超过10天
	if midnight.Sub(*ts) > 11*time.Hour*24 {
		t.Log("超过11天")
	} else {
		t.Log("未超过11天")
	}
}

func Test_clearLog(t *testing.T) {
	clearLog()
}

// 动态level 测试
func TestNewAtomicLevelAt(t *testing.T) {

	//SetEnv(LOG_PRO)
	//SetLog(ENV_DEBUG)
	SetLog(ENV_WARN)
	SetLog(ENV_DEBUG)
	F().Warn("test_warn")
	F().Info("test_info")
	F().Debug("test_debug")
	F().Error("test_error")

	time.Sleep(3 * time.Second)
}

func TestMangets(t *testing.T) {
	mgr := NewManager()
	mgr.SetLog(
		ENV_WARN,
		WithRotationTime(12),
	)

	svcLog := mgr.F("payment")
	svcLog.Warn("order pending")
}

// 测试动态修改日志级别功能
func TestDynamicDebugLevel(t *testing.T) {
	t.Log("=== 测试动态修改日志级别 ===")

	// 1. 先设置为 PRO 环境（不输出到终端）
	t.Log("步骤1: 设置为 PRO 环境")
	SetLog(ENV_PRO)
	
	// 获取一个 logger 并输出日志
	t.Log("步骤2: 在 PRO 环境下输出日志（应该只写文件，不输出到终端）")
	F().Info("PRO 环境 - Info 日志")
	F().Debug("PRO 环境 - Debug 日志（应该不会记录，因为级别是 Info）")

	time.Sleep(1 * time.Second)

	// 2. 动态切换到 Debug 级别
	t.Log("步骤3: 调用 SetDebugLevel() 动态切换到 Debug 级别")
	SetDebugLevel()

	// 3. 再次输出日志，验证是否能输出到终端
	t.Log("步骤4: 切换后输出日志（应该输出到终端）")
	F().Info("切换后 - Info 日志（应该输出到终端）")
	F().Debug("切换后 - Debug 日志（应该输出到终端）")
	F().Warn("切换后 - Warn 日志（应该输出到终端）")

	time.Sleep(1 * time.Second)

	t.Log("=== 测试完成，请检查终端是否看到'切换后'的日志 ===")
}

// 测试使用独立 Manager 的动态级别切换
func TestManagerDynamicDebugLevel(t *testing.T) {
	t.Log("=== 测试 Manager 动态修改日志级别 ===")

	// 创建独立的 Manager
	mgr := NewManager()

	// 1. 设置为 PRO 环境
	t.Log("步骤1: 设置为 PRO 环境")
	mgr.SetLog(ENV_PRO)

	t.Log("步骤2: 在 PRO 环境下输出日志")
	mgr.F("test_manager").Info("Manager PRO 环境 - Info 日志")
	mgr.F("test_manager").Debug("Manager PRO 环境 - Debug 日志")

	time.Sleep(1 * time.Second)

	// 2. 动态切换到 Debug 级别
	t.Log("步骤3: 调用 SetDebugLevel() 动态切换")
	mgr.SetDebugLevel()

	// 3. 验证切换效果
	t.Log("步骤4: 切换后输出日志（应该输出到终端）")
	mgr.F("test_manager").Info("Manager 切换后 - Info 日志")
	mgr.F("test_manager").Debug("Manager 切换后 - Debug 日志")
	mgr.F("test_manager").Warn("Manager 切换后 - Warn 日志")

	time.Sleep(1 * time.Second)

	t.Log("=== 测试完成 ===")
}

// 测试所有日志级别的快捷设置方法
func TestAllLevelSetters(t *testing.T) {
	t.Log("=== 测试所有日志级别快捷方法 ===")

	// 测试全局方法
	t.Log("测试全局方法:")

	SetDebugLevel()
	t.Log("- SetDebugLevel() 完成")
	F().Debug("Debug 级别日志")
	F().Info("Info 级别日志")

	SetInfoLevel()
	t.Log("- SetInfoLevel() 完成")
	F().Info("Info 级别日志")

	SetWarnLevel()
	t.Log("- SetWarnLevel() 完成")
	F().Warn("Warn 级别日志")

	SetErrorLevel()
	t.Log("- SetErrorLevel() 完成")
	F().Error("Error 级别日志")

	// DPanic, Panic, Fatal 在测试中不实际调用，以免影响测试进程
	SetDPanicLevel()
	t.Log("- SetDPanicLevel() 完成")

	// 测试 Manager 实例方法
	t.Log("\n测试 Manager 实例方法:")
	mgr := NewManager()

	mgr.SetDebugLevel()
	t.Log("- Manager.SetDebugLevel() 完成")
	mgr.F("test").Debug("Manager Debug 日志")

	mgr.SetInfoLevel()
	t.Log("- Manager.SetInfoLevel() 完成")
	mgr.F("test").Info("Manager Info 日志")

	mgr.SetWarnLevel()
	t.Log("- Manager.SetWarnLevel() 完成")
	mgr.F("test").Warn("Manager Warn 日志")

	mgr.SetErrorLevel()
	t.Log("- Manager.SetErrorLevel() 完成")
	mgr.F("test").Error("Manager Error 日志")

	mgr.SetDPanicLevel()
	t.Log("- Manager.SetDPanicLevel() 完成")

	mgr.SetPanicLevel()
	t.Log("- Manager.SetPanicLevel() 完成")

	mgr.SetFatalLevel()
	t.Log("- Manager.SetFatalLevel() 完成")

	t.Log("=== 所有级别设置方法测试通过 ===")
}

// 并发压测：测试在高并发场景下动态切换日志级别
func TestConcurrentDynamicLevel(t *testing.T) {
	t.Log("=== 开始并发压测 ===")

	// 初始化为 PRO 环境
	SetLog(ENV_PRO)

	// 并发参数
	numGoroutines := 100     // 并发goroutine数量
	operationsPerGoroutine := 1000  // 每个goroutine的操作次数
	numLevelSwitches := 50   // 日志级别切换次数

	// 用于同步的channel
	done := make(chan bool, numGoroutines+1)

	// 启动多个goroutine持续写日志
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 使用不同的logger名称
				loggerName := "test_logger"
				if id%3 == 0 {
					loggerName = "logger_a"
				} else if id%3 == 1 {
					loggerName = "logger_b"
				}

				// 输出不同级别的日志
				switch j % 4 {
				case 0:
					F(loggerName).Infof("goroutine-%d iteration-%d", id, j)
				case 1:
					F(loggerName).Debugf("goroutine-%d iteration-%d", id, j)
				case 2:
					F(loggerName).Warnf("goroutine-%d iteration-%d", id, j)
				case 3:
					F(loggerName).Errorf("goroutine-%d iteration-%d", id, j)
				}
			}
		}(i)
	}

	// 启动一个goroutine持续切换日志级别
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Level switcher panicked: %v", r)
			}
			done <- true
		}()

		for i := 0; i < numLevelSwitches; i++ {
			time.Sleep(10 * time.Millisecond)

			// 在不同的日志级别之间切换
			switch i % 4 {
			case 0:
				SetDebugLevel()
			case 1:
				SetLog(ENV_INFO)
			case 2:
				SetLog(ENV_WARN)
			case 3:
				SetLog(ENV_PRO)
			}
		}
	}()

	// 等待所有goroutine完成
	for i := 0; i < numGoroutines+1; i++ {
		<-done
	}

	t.Logf("=== 压测完成: %d个goroutine, 每个%d次操作, %d次级别切换 ===",
		numGoroutines, operationsPerGoroutine, numLevelSwitches)
}

// 测试Manager的并发安全性
func TestConcurrentManagerDynamicLevel(t *testing.T) {
	t.Log("=== 开始Manager并发压测 ===")

	mgr := NewManager()
	mgr.SetLog(ENV_PRO)

	numGoroutines := 50
	operationsPerGoroutine := 500
	numLevelSwitches := 30

	done := make(chan bool, numGoroutines+1)

	// 多个goroutine写日志
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Manager goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			for j := 0; j < operationsPerGoroutine; j++ {
				mgr.F("concurrent_test").Infof("mgr-goroutine-%d iteration-%d", id, j)
				mgr.F("concurrent_test").Debugf("mgr-goroutine-%d iteration-%d", id, j)
			}
		}(i)
	}

	// 一个goroutine切换级别
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Manager level switcher panicked: %v", r)
			}
			done <- true
		}()

		for i := 0; i < numLevelSwitches; i++ {
			time.Sleep(10 * time.Millisecond)

			if i%2 == 0 {
				mgr.SetDebugLevel()
			} else {
				mgr.SetLog(ENV_WARN)
			}
		}
	}()

	// 等待完成
	for i := 0; i < numGoroutines+1; i++ {
		<-done
	}

	t.Logf("=== Manager压测完成 ===")
}

// Benchmark: 测试动态级别切换的性能
func BenchmarkDynamicLevelSwitch(b *testing.B) {
	SetLog(ENV_PRO)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			SetDebugLevel()
		} else {
			SetLog(ENV_INFO)
		}
	}
}

// Benchmark: 测试日志写入性能
func BenchmarkLogging(b *testing.B) {
	SetLog(ENV_PRO)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		F().Infof("benchmark test %d", i)
	}
}

// Benchmark: 并发写入性能
func BenchmarkConcurrentLogging(b *testing.B) {
	SetLog(ENV_PRO)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			F().Infof("concurrent benchmark %d", i)
			i++
		}
	})
}
