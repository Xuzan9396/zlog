package zlog

import (
	"testing"
	"time"
)

// TestConsoleOnlyMode 测试仅输出到终端模式
func TestConsoleOnlyMode(t *testing.T) {
	t.Log("=== 测试仅输出到终端模式 ===")

	// 1. 先测试默认模式（写文件）
	t.Log("步骤1: 默认模式（写文件）")
	SetLog(ENV_DEBUG)
	F().Info("默认模式 - 应该写入文件")
	F("test_service").Info("默认模式 - test_service 日志")

	time.Sleep(500 * time.Millisecond)

	// 2. 切换到仅终端模式
	t.Log("步骤2: 切换到仅终端输出模式")
	SetConsoleOnly(true)
	F().Info("仅终端模式 - 不应该写入文件")
	F("api").Info("仅终端模式 - api 日志")
	F("v1").Warn("仅终端模式 - v1 警告")

	time.Sleep(500 * time.Millisecond)

	// 3. 切换回文件模式
	t.Log("步骤3: 切换回文件模式")
	SetConsoleOnly(false)
	F().Info("恢复文件模式 - 应该写入文件")

	time.Sleep(500 * time.Millisecond)

	t.Log("=== 测试完成 ===")
}

// TestConsoleOnlyWithOptions 测试通过选项设置仅终端模式
func TestConsoleOnlyWithOptions(t *testing.T) {
	t.Log("=== 测试通过选项设置仅终端模式 ===")

	SetLog(ENV_DEBUG, WithConsoleOnly(true))
	F().Info("使用 WithConsoleOnly 选项")
	F("service").Info("service 日志")

	time.Sleep(500 * time.Millisecond)

	t.Log("=== 测试完成 ===")
}

// TestManagerConsoleOnly 测试 Manager 实例的仅终端模式
func TestManagerConsoleOnly(t *testing.T) {
	t.Log("=== 测试 Manager 实例的仅终端模式 ===")

	mgr := NewManager()
	mgr.SetLog(ENV_DEBUG)

	t.Log("步骤1: Manager 默认模式")
	mgr.F().Info("Manager 默认模式")
	mgr.F("payment").Info("Manager payment 日志")

	time.Sleep(500 * time.Millisecond)

	t.Log("步骤2: Manager 切换到仅终端模式")
	mgr.SetConsoleOnly(true)
	mgr.F().Info("Manager 仅终端模式")
	mgr.F("order").Warn("Manager order 警告")

	time.Sleep(500 * time.Millisecond)

	t.Log("=== 测试完成 ===")
}

// TestFieldName 测试 f 字段功能
func TestFieldName(t *testing.T) {
	t.Log("=== 测试 f 字段功能 ===")

	SetLog(ENV_DEBUG, WithConsoleOnly(true))

	t.Log("默认 logger（应该显示 f=log）:")
	F().Info("这是默认 logger，应该显示 f=log")

	t.Log("\n指定名称的 logger（应该有 f 字段）:")
	F("v1").Info("这是 v1，应该有 f=v1")
	F("api").Warn("这是 api，应该有 f=api")
	F("service").Error("这是 service，应该有 f=service")

	time.Sleep(500 * time.Millisecond)

	t.Log("=== 测试完成，请检查上面的输出 ===")
}

// TestDynamicSwitchConsoleOnly 测试动态切换终端模式
func TestDynamicSwitchConsoleOnly(t *testing.T) {
	t.Log("=== 测试动态切换终端模式 ===")

	SetLog(ENV_DEBUG)

	t.Log("步骤1: 文件模式")
	F("app").Info("文件模式日志")

	SetConsoleOnly(true)
	t.Log("步骤2: 切换到终端模式")
	F("app").Info("终端模式日志")

	SetConsoleOnly(false)
	t.Log("步骤3: 切换回文件模式")
	F("app").Info("恢复文件模式日志")

	time.Sleep(500 * time.Millisecond)

	t.Log("=== 测试完成 ===")
}
