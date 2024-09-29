package zlog

import (
	"strconv"
	"testing"
	"time"
)

func Test_getLogDate(t *testing.T) {
	logFileName := "new_logs_info2024-03-16.log"

	_, ts, err := getLogDate(logFileName)
	if err != nil {
		t.Error(err)
		return
	}
	var LOC, _ = time.LoadLocation("Local")
	now := time.Now() // 确保now使用和ts相同的时区

	// 获取当前时间的年、月、日部分，并将小时、分钟、秒和纳秒设置为0，以便进行日期比较
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, LOC)

	//// 计算两个日期之间的差异天数，忽略时间部分
	//daysDiff := midnight.Sub(*ts).Hours() / 24
	//t.Log(ts.Unix(), midnight.Unix(), daysDiff)
	//
	//t.Logf("%s %v %v", logFileName, ts.Format("2006-01-02"), daysDiff)

	// 判断ts与当前时间的差值是否超过10天
	if midnight.Sub(*ts) > 11*time.Hour*24 {
		t.Log("超过11天")
	} else {
		t.Log("未超过11天")
	}
}

func Test_clearLog(t *testing.T) {
	clearLog()
}

// 动态level 测试
func TestNewAtomicLevelAt(t *testing.T) {
	i := 1

	//SetEnv(LOG_PRO)
	//SetLog(ENV_DEBUG)
	for {
		F().Debugf("%s", "测试下好不好"+strconv.Itoa(i))
		if i >= 200 {
			//atomicLevel.SetLevel(zapcore.DebugLevel)
			SetDebugLevel()
		}
		if i >= 1000 {
			break
		}

		i++
	}
	time.Sleep(5 * time.Second)
}
