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

// init 不再提前创建日志目录，而是在实际需要写入文件时才创建。
// 注意：清理由后台任务定期执行。
func init() {
	// 不再提前创建目录，延迟到实际需要写入文件时
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
// 使用程序启动时的工作目录（运行目录）作为根目录。
func defaultLogsDir() string {
	// 获取当前工作目录（程序运行目录）
	cwd, err := os.Getwd()
	if err != nil {
		// 如果获取失败，使用相对路径作为兜底
		if runtime.GOOS == "windows" {
			return ".\\logs"
		}
		return "./logs"
	}

	// 返回工作目录下的 logs 子目录（绝对路径）
	return filepath.Join(cwd, "logs")
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
