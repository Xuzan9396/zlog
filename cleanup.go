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

	// 第二遍：清理所有已失效的软链接（避免因处理顺序导致遗留）
	cleanInvalidSymlinks(logPath)
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

		// 如果仍有同前缀的实际日志文件（可能是硬链接或新文件）则保留
		prefix := loggerPrefixFromLink(entry.Name())
		if hasRegularLogWithPrefix(dir, prefix) {
			continue
		}

		// 目标缺失且无同前缀文件，删除软链接
		_ = os.Remove(fullPath)
	}
}

// loggerPrefixFromLink 提取软链接名中的 logger 前缀（例如 log_info.log -> log）。
func loggerPrefixFromLink(linkName string) string {
	base := filepath.Base(linkName)
	if idx := strings.Index(base, "_"); idx > 0 {
		return base[:idx]
	}
	if idx := strings.Index(base, "."); idx > 0 {
		return base[:idx]
	}
	return base
}

// hasRegularLogWithPrefix 检查目录下是否存在给定前缀的常规日志文件。
func hasRegularLogWithPrefix(dir, prefix string) bool {
	if prefix == "" {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.Type().IsRegular() && strings.HasPrefix(e.Name(), prefix) && strings.Contains(e.Name(), ".log") {
			return true
		}
	}
	return false
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
