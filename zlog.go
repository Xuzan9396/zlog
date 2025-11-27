package zlog

import "go.uber.org/zap"

// F 返回指定名称的 SugaredLogger，默认使用 sign 通道。
// 可选参数第二位用于向外暴露封装层时调整调用栈深度。
func F(fileNameArr ...string) *zap.SugaredLogger {
	return getDefaultManager().Logger(fileNameArr...)
}

// Sync 主动刷新某个 logger 的缓冲区，确保日志落盘。
func Sync(fileName string) error {
	return getDefaultManager().Sync(fileName)
}

// SetDebugLevel 将全局等级动态调整为 Debug，便于排查在线问题。
func SetDebugLevel() {
	getDefaultManager().SetDebugLevel()
}

// SetInfoLevel 将全局等级动态调整为 Info。
func SetInfoLevel() {
	getDefaultManager().SetInfoLevel()
}

// SetWarnLevel 将全局等级动态调整为 Warn。
func SetWarnLevel() {
	getDefaultManager().SetWarnLevel()
}

// SetErrorLevel 将全局等级动态调整为 Error。
func SetErrorLevel() {
	getDefaultManager().SetErrorLevel()
}

// SetDPanicLevel 将全局等级动态调整为 DPanic。
func SetDPanicLevel() {
	getDefaultManager().SetDPanicLevel()
}

// SetPanicLevel 将全局等级动态调整为 Panic。
func SetPanicLevel() {
	getDefaultManager().SetPanicLevel()
}

// SetFatalLevel 将全局等级动态调整为 Fatal。
func SetFatalLevel() {
	getDefaultManager().SetFatalLevel()
}

// SetConsoleOnly 动态设置全局仅输出到终端模式。
func SetConsoleOnly(consoleOnly bool) {
	getDefaultManager().SetConsoleOnly(consoleOnly)
}
