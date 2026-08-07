package zlog

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 包级正则表达式，避免重复编译
var logDateRegex = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\.log(?:\.\d+)?$`)

// file-rotatelogs 的 _lock/_symlink 只应在一次轮转期间短暂存在；
// 超过一小时仍未完成可判定为进程异常退出后的残留，同时给慢磁盘留出足够安全余量。
const staleRotationArtifactAge = time.Hour

// clearLog 清理超出保留时长的旧日志文件，同时处理软链接。
// 改进版本：扫描整个日志目录，清理所有过期文件（包括不再使用的 logger）
func clearLog() {
	clearLogWithConfig(nil)
}

// clearLogWithConfig 使用指定配置清理日志，如果 cfg 为 nil，使用全局配置
func clearLogWithConfig(cfg *Config) {
	var maxAge int
	if cfg == nil {
		// 使用全局配置
		globalCfg := getConfig()
		maxAge = globalCfg.WithMaxAge
	} else {
		maxAge = cfg.WithMaxAge
	}

	logPath := logDir()

	// 扫描日志目录
	entries, err := os.ReadDir(logPath)
	if err != nil {
		return
	}

	now := time.Now().In(location)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	expireWindow := time.Duration(maxAge) * time.Hour

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		fullPath := filepath.Join(logPath, fileName)

		// 只处理日志文件
		if !isLogFile(fileName) {
			continue
		}

		// 提取日期
		ts := extractLogDate(fileName)
		if ts == nil {
			// 没有日期的文件（可能是软链接）
			if isSymlink(fullPath) {
				// 检查软链接是否有效
				if !isValidSymlink(fullPath) {
					_ = os.Remove(fullPath)
				}
			}
			continue
		}

		// 删除过期文件
		if midnight.Sub(*ts) > expireWindow {
			_ = os.Remove(fullPath)
		}
	}

	// rotatelogs 自身会忽略 _lock/_symlink；进程在创建临时文件后异常退出时，
	// 必须由 zlog 清理，否则文件监听器可能因为断开的软链接而无法启动。
	cleanRotationArtifacts(logPath, now)

	// 第二遍：清理所有已失效的软链接（避免因处理顺序导致遗留）
	cleanInvalidSymlinks(logPath)
}

// cleanRotationArtifacts 清理滚动日志异常中断后遗留的临时文件。
// 例: 断开的 foo.log_symlink 立即删除；仍有效但超过1小时的临时软链和旧 lock 也会删除。
func cleanRotationArtifacts(dir string, now time.Time) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		fullPath := filepath.Join(dir, name)

		switch {
		case strings.HasSuffix(name, "_symlink") && entry.Type()&os.ModeSymlink != 0:
			info, statErr := os.Lstat(fullPath)
			if statErr != nil {
				continue
			}
			// 断链不会再被 rotatelogs 使用，应立即删除；有效临时链只在超过安全窗口后清理。
			if !isValidSymlink(fullPath) || now.Sub(info.ModTime()) > staleRotationArtifactAge {
				_ = os.Remove(fullPath)
			}
		case strings.HasSuffix(name, "_lock") && entry.Type().IsRegular():
			info, statErr := entry.Info()
			if statErr != nil {
				continue
			}
			if now.Sub(info.ModTime()) > staleRotationArtifactAge {
				_ = os.Remove(fullPath)
			}
		}
	}
}

// isLogFile 判断是否是日志文件
func isLogFile(fileName string) bool {
	// 以 .log 结尾，或包含日期模式
	return strings.Contains(fileName, ".log")
}

// isSymlink 判断是否是软链接
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// isValidSymlink 判断软链接是否有效
func isValidSymlink(path string) bool {
	_, err := os.Stat(path) // 使用 Stat 会跟随软链接
	return err == nil
}

// cleanInvalidSymlinks 删除日志目录中指向不存在目标的软链接。
func cleanInvalidSymlinks(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		fullPath := filepath.Join(dir, entry.Name())
		// 目标正常存在则保留
		if isValidSymlink(fullPath) {
			continue
		}

		// 断开的软链接本身已不可用；同系列存在其他日志文件也不能让该链接恢复有效。
		_ = os.Remove(fullPath)
	}
}

// extractLogDate 从文件名提取日期
func extractLogDate(fileName string) *time.Time {
	match := logDateRegex.FindStringSubmatch(fileName)
	if match == nil || len(match) < 2 {
		return nil
	}

	const layout = "2006-01-02"
	ts, err := time.ParseInLocation(layout, match[1], location)
	if err != nil {
		return nil
	}

	return &ts
}

// CleanupTask 后台清理任务管理器
type CleanupTask struct {
	interval time.Duration
	ticker   *time.Ticker
	stopChan chan struct{}
	running  atomic.Bool
	mu       sync.Mutex
}

// newCleanupTask 创建新的清理任务
func newCleanupTask(interval time.Duration) *CleanupTask {
	return &CleanupTask{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start 启动后台清理任务
func (t *CleanupTask) Start() {
	if !t.running.CompareAndSwap(false, true) {
		return // 已经在运行
	}

	go t.run()
}

// run 执行清理任务循环
func (t *CleanupTask) run() {
	defer t.running.Store(false)

	t.mu.Lock()
	t.ticker = time.NewTicker(t.interval)
	t.mu.Unlock()

	defer t.ticker.Stop()

	for {
		select {
		case <-t.ticker.C:
			clearLog()
		case <-t.stopChan:
			return
		}
	}
}

// Stop 停止清理任务
func (t *CleanupTask) Stop() {
	if !t.running.Load() {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	select {
	case <-t.stopChan:
		// 已经关闭
	default:
		close(t.stopChan)
	}
}

// IsRunning 返回清理任务是否正在运行
func (t *CleanupTask) IsRunning() bool {
	return t.running.Load()
}

// getLogDate 解析日志文件名中的日期部分，并返回文件前缀。
func getLogDate(logFileName string) (prefix string, logDate *time.Time, err error) {
	re := regexp.MustCompile(`^(.*?)(\d{4}-\d{2}-\d{2})\.log(?:\.\d+)?$`)
	match := re.FindStringSubmatch(logFileName)
	if match == nil || len(match) != 3 {
		return "", nil, errors.New("no date found in string")
	}

	const layout = "2006-01-02"
	ts, err := time.ParseInLocation(layout, match[2], location)
	if err != nil {
		return match[1], &ts, err
	}

	return match[1], &ts, nil
}
