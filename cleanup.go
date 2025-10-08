package zlog

import (
	"errors"
	"path/filepath"
	"regexp"
	"time"
)

// clearLog 清理超出保留时长的旧日志文件，同时处理软链接。
func clearLog() {
	cfg := getConfig()
	pattern := filepath.Join(logDir(), "*20*.log*")
	initialFiles, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	expired := make(map[string]struct{})
	retained := make(map[string]struct{})

	now := time.Now().In(location)
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	expireWindow := time.Duration(cfg.WithMaxAge+24) * time.Hour

	for _, logFileName := range initialFiles {
		prefixName, ts, err := getLogDate(logFileName)
		if err != nil || ts == nil {
			continue
		}

		if midnight.Sub(*ts) > expireWindow {
			_ = removeFile(logFileName)
			if prefixName != "" {
				expired[prefixName+".log"] = struct{}{}
			}
		} else if prefixName != "" {
			retained[prefixName+".log"] = struct{}{}
		}
	}

	for name := range expired {
		if _, ok := retained[name]; ok {
			continue
		}
		_ = removeFile(name)
	}
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
