package zlog

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetWorkingDirectory 测试获取工作目录功能
func TestGetWorkingDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("获取工作目录失败: %v", err)
		return
	}
	t.Logf("当前工作目录: %s", cwd)

	if cwd == "" {
		t.Error("工作目录为空")
	}

	if !filepath.IsAbs(cwd) {
		t.Error("工作目录不是绝对路径")
	}
}

// TestDefaultLogsDir 测试默认日志目录
func TestDefaultLogsDir(t *testing.T) {
	logsDir := defaultLogsDir()
	t.Logf("默认日志目录: %s", logsDir)

	if logsDir == "" {
		t.Error("日志目录为空")
		return
	}

	// 检查是否是绝对路径
	if filepath.IsAbs(logsDir) {
		t.Logf("✓ 日志目录是绝对路径")
	} else {
		t.Logf("⚠ 日志目录是相对路径（可能未找到项目根目录）")
	}

	// 检查路径是否包含 "logs"
	if filepath.Base(logsDir) == "logs" {
		t.Logf("✓ 日志目录名称正确")
	}
}

// TestLogDir 测试当前使用的日志目录
func TestLogDir(t *testing.T) {
	dir := logDir()
	t.Logf("当前日志目录: %s", dir)

	if dir == "" {
		t.Error("日志目录为空")
	}
}

// TestLogsDirectoryInWorkingDirectory 测试日志目录是否在工作目录下
func TestLogsDirectoryInWorkingDirectory(t *testing.T) {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("当前工作目录: %s", cwd)

	// 获取日志目录
	logsDir := logDir()
	t.Logf("日志目录: %s", logsDir)

	// 验证日志目录是否在工作目录下
	expectedLogsDir := filepath.Join(cwd, "logs")
	if logsDir == expectedLogsDir {
		t.Logf("✓ 日志目录正确设置在工作目录下")
	} else {
		t.Errorf("日志目录不在工作目录下。期望: %s, 实际: %s", expectedLogsDir, logsDir)
	}
}
