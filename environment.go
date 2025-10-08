package zlog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// 包级变量初始化在早期完成，确保日志目录、时区可用。
var (
	location = loadLocation()
	logsDir  = defaultLogsDir()
)

// init 负责创建日志目录并立即进行一次历史清理。
func init() {
	if err := ensureDir(logsDir); err != nil {
		fmt.Fprintf(os.Stderr, "zlog: ensure log dir failed: %v\n", err)
	}
	clearLog()
}

// logDir 返回当前使用的日志根目录。
func logDir() string {
	return logsDir
}

// loadLocation 加载本地时区，失败时兜底为系统默认。
func loadLocation() *time.Location {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return time.Local
	}
	return loc
}

// defaultLogsDir 根据系统选择默认日志目录。
func defaultLogsDir() string {
	if runtime.GOOS == "windows" {
		return ".\\logs"
	}
	return "./logs"
}

// ensureDir 确保目录存在且为文件夹。
func ensureDir(dirName string) error {
	info, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		return os.MkdirAll(dirName, 0o755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", dirName)
	}
	return nil
}

func removeFile(path string) error {
	return os.Remove(path)
}

// logFilePath 在日志目录下拼接出具体文件路径。
func logFilePath(format string, args ...interface{}) string {
	return filepath.Join(logDir(), fmt.Sprintf(format, args...))
}
