package zlog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCleanupTask 测试后台清理任务
func TestCleanupTask(t *testing.T) {
	t.Log("=== 测试后台清理任务 ===")

	SetLog(ENV_DEBUG, WithMaxAge(24*8)) // 保留 8 天的

	// 创建测试 Manager，配置短间隔便于测试
	mgr := NewManager(
		WithAutoCleanup(true),
		WithCleanupInterval(2*time.Second), // 2秒清理一次
		WithMaxAge(1),                      // 保留1小时
	)

	// 等待清理任务启动（延迟启动）
	time.Sleep(200 * time.Millisecond)

	// 验证清理任务已启动
	if !mgr.IsCleanupRunning() {
		t.Error("清理任务应该已经启动")
	}

	t.Log("清理任务已启动")

	// 等待一段时间
	time.Sleep(3 * time.Second)

	// 停止清理任务
	mgr.StopCleanupTask()

	time.Sleep(500 * time.Millisecond)

	// 验证清理任务已停止
	if mgr.IsCleanupRunning() {
		t.Error("清理任务应该已经停止")
	}

	t.Log("清理任务已停止")
	t.Log("=== 测试完成 ===")
}

// TestCleanupRemovesBrokenSymlink 验证清理会删除失效软链接
func TestCleanupRemovesBrokenSymlink(t *testing.T) {
	t.Log("=== 测试清理失效软链接 ===")

	// 临时日志目录
	tmpDir := filepath.Join(os.TempDir(), "zlog_test_broken_symlink")
	_ = os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}
	// 覆盖全局日志目录
	origDir := logsDir
	logsDir = tmpDir
	defer func() {
		logsDir = origDir
		_ = os.RemoveAll(tmpDir)
	}()

	// 创建过期文件和指向它的软链接
	expiredName := "foo_info2020-01-01.log"
	expiredPath := filepath.Join(tmpDir, expiredName)
	if err := os.WriteFile(expiredPath, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	linkName := filepath.Join(tmpDir, "foo_info.log")
	if err := os.Symlink(expiredName, linkName); err != nil {
		t.Fatal(err)
	}

	// 使用 1 小时保留期触发删除
	clearLogWithConfig(&Config{WithMaxAge: 1})

	// 软链接应被移除
	if _, err := os.Lstat(linkName); err == nil {
		t.Fatalf("软链接未被删除: %s", linkName)
	}
	t.Log("失效软链接已删除")

	// 再创建一个软链但保留同前缀的真实文件，验证不会误删
	validDate := time.Now().Format("2006-01-02")
	validFile := filepath.Join(tmpDir, "foo_info"+validDate+".log")
	if err := os.WriteFile(validFile, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Base(validFile), linkName); err != nil {
		t.Fatal(err)
	}

	clearLogWithConfig(&Config{WithMaxAge: 1})

	if _, err := os.Lstat(linkName); err != nil {
		t.Fatalf("软链接被误删: %v", err)
	}
	t.Log("存在同前缀文件时软链接被保留")
}

// TestManualCleanup 测试手动清理
func TestManualCleanup(t *testing.T) {
	t.Log("=== 测试手动清理 ===")

	// 创建测试日志目录
	testLogDir := filepath.Join(os.TempDir(), "zlog_test_cleanup")
	defer os.RemoveAll(testLogDir)

	if err := os.MkdirAll(testLogDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建一些过期的测试日志文件
	oldDate := time.Now().AddDate(0, 0, -15).Format("2006-01-02")
	oldFile := filepath.Join(testLogDir, "test_info"+oldDate+".log")
	if err := os.WriteFile(oldFile, []byte("old log"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Logf("创建了过期测试文件: %s", oldFile)

	// 手动触发清理
	CleanupLogs()

	t.Log("手动清理已触发")
	t.Log("=== 测试完成 ===")
}

// TestDisableAutoCleanup 测试禁用自动清理
func TestDisableAutoCleanup(t *testing.T) {
	t.Log("=== 测试禁用自动清理 ===")

	// 创建禁用自动清理的 Manager
	mgr := NewManager(WithAutoCleanup(false))

	// 验证清理任务未启动
	if mgr.IsCleanupRunning() {
		t.Error("清理任务不应该启动")
	}

	t.Log("自动清理已禁用，清理任务未启动")
	t.Log("=== 测试完成 ===")
}

// TestCleanupIntervalChange 测试动态修改清理间隔
func TestCleanupIntervalChange(t *testing.T) {
	t.Log("=== 测试动态修改清理间隔 ===")

	// 创建 Manager
	mgr := NewManager(
		WithAutoCleanup(true),
		WithCleanupInterval(5*time.Second),
	)

	// 等待清理任务启动
	time.Sleep(200 * time.Millisecond)

	if !mgr.IsCleanupRunning() {
		t.Error("清理任务应该已经启动")
	}

	t.Log("清理任务已启动（间隔5秒）")

	// 修改配置，改变清理间隔
	mgr.SetLog(ENV_INFO,
		WithAutoCleanup(true),
		WithCleanupInterval(3*time.Second), // 改为3秒
	)

	time.Sleep(500 * time.Millisecond)

	if !mgr.IsCleanupRunning() {
		t.Error("清理任务应该仍在运行")
	}

	t.Log("清理间隔已更新（改为3秒）")

	// 清理
	mgr.StopCleanupTask()

	t.Log("=== 测试完成 ===")
}

// TestGlobalCleanupAPIs 测试全局清理 API
func TestGlobalCleanupAPIs(t *testing.T) {
	t.Log("=== 测试全局清理 API ===")

	// 检查默认管理器的清理任务状态
	if !IsCleanupRunning() {
		t.Log("默认管理器的清理任务未运行（可能已在其他测试中停止）")
	} else {
		t.Log("默认管理器的清理任务正在运行")
	}

	// 手动触发全局清理
	CleanupLogs()
	t.Log("全局手动清理已触发")

	t.Log("=== 测试完成 ===")
}

// TestExtractLogDate 测试日期提取函数
func TestExtractLogDate(t *testing.T) {
	t.Log("=== 测试日期提取函数 ===")

	testCases := []struct {
		fileName string
		expected bool
	}{
		{"test_info2025-01-15.log", true},
		{"service_a_info2025-12-31.log", true},
		{"log_error2024-06-20.log.1", true},
		{"test_info.log", false},
		{"random_file.txt", false},
		{"2025-01-15.log", true},
	}

	for _, tc := range testCases {
		ts := extractLogDate(tc.fileName)
		hasDate := ts != nil

		if hasDate != tc.expected {
			t.Errorf("文件 %s: 期望有日期=%v, 实际=%v", tc.fileName, tc.expected, hasDate)
		} else {
			if hasDate {
				t.Logf("✓ %s -> %s", tc.fileName, ts.Format("2006-01-02"))
			} else {
				t.Logf("✓ %s -> 无日期", tc.fileName)
			}
		}
	}

	t.Log("=== 测试完成 ===")
}

// TestIsLogFile 测试日志文件判断
func TestIsLogFile(t *testing.T) {
	t.Log("=== 测试日志文件判断 ===")

	testCases := []struct {
		fileName string
		expected bool
	}{
		{"test.log", true},
		{"test2025-01-15.log", true},
		{"test.log.1", true},
		{"test.txt", false},
		{"readme.md", false},
	}

	for _, tc := range testCases {
		result := isLogFile(tc.fileName)
		if result != tc.expected {
			t.Errorf("文件 %s: 期望=%v, 实际=%v", tc.fileName, tc.expected, result)
		} else {
			t.Logf("✓ %s -> %v", tc.fileName, result)
		}
	}

	t.Log("=== 测试完成 ===")
}
