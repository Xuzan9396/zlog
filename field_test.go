package zlog

import (
	"testing"
)

// TestFieldOutput 测试 f 字段的输出行为
func TestFieldOutput(t *testing.T) {
	t.Log("=== 测试 f 字段输出行为 ===")

	// 测试仅终端模式：f 字段固定显示
	t.Log("\n1. 仅终端模式测试:")
	SetLog(ENV_DEBUG, WithConsoleOnly(true))

	t.Log("默认 logger（应显示 f=log）:")
	F().Info("默认 logger 测试")

	t.Log("\n指定名称 logger（应显示 f=对应名称）:")
	F("api").Info("api logger 测试")
	F("v1").Warn("v1 logger 测试")
	F("payment").Error("payment logger 测试")

	// 测试文件+终端模式
	t.Log("\n2. 文件+终端模式测试（ENV_DEBUG）:")
	SetLog(ENV_DEBUG, WithConsoleOnly(false))

	t.Log("终端应显示 f 字段，文件不应显示 f 字段")
	F().Info("默认 logger - 文件模式")
	F("service").Info("service logger - 文件模式")

	// 测试纯文件模式（无终端输出）
	t.Log("\n3. 纯文件模式测试（ENV_INFO）:")
	SetLog(ENV_INFO, WithConsoleOnly(false))

	t.Log("应该只写文件，不输出到终端，文件中无 f 字段")
	F().Info("这条日志只写文件")
	F("worker").Info("worker 只写文件")

	t.Log("\n=== 测试完成，请检查: ===")
	t.Log("1. 终端输出都有 f 字段")
	t.Log("2. 默认 logger 的 f=log")
	t.Log("3. 文件中没有 f 字段（需要手动检查日志文件）")
}

// TestDefaultLoggerFieldName 测试默认 logger 的 f 字段值
func TestDefaultLoggerFieldName(t *testing.T) {
	t.Log("=== 测试默认 logger 的 f 字段 ===")

	SetLog(ENV_DEBUG, WithConsoleOnly(true))

	t.Log("默认 logger 输出（应显示 f=log）:")
	F().Info("测试默认 logger")
	F().Debug("测试默认 logger debug")
	F().Warn("测试默认 logger warn")

	t.Log("\n=== 测试完成，请确认终端输出 f=log ===")
}

// TestNamedLoggerFieldName 测试指定名称 logger 的 f 字段值
func TestNamedLoggerFieldName(t *testing.T) {
	t.Log("=== 测试指定名称 logger 的 f 字段 ===")

	SetLog(ENV_DEBUG, WithConsoleOnly(true))

	t.Log("不同名称的 logger 输出:")
	F("api").Info("api logger - 应显示 f=api")
	F("worker").Info("worker logger - 应显示 f=worker")
	F("scheduler").Info("scheduler logger - 应显示 f=scheduler")

	t.Log("\n=== 测试完成，请确认 f 字段显示正确的名称 ===")
}
