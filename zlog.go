package zlog

import "go.uber.org/zap"

// F 返回指定名称的 SugaredLogger，默认使用 sign 通道。
// 可选参数第二位用于向外暴露封装层时调整调用栈深度。
func F(fileNameArr ...string) *zap.SugaredLogger {
	return defaultManager.Logger(fileNameArr...)
}

// Sync 主动刷新某个 logger 的缓冲区，确保日志落盘。
func Sync(fileName string) error {
	return defaultManager.Sync(fileName)
}

// SetDebugLevel 将全局等级动态调整为 Debug，便于排查在线问题。
func SetDebugLevel() {
	defaultManager.SetDebugLevel()
}
