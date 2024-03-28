package zlog

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type Env string
type EnvDate string // 时间格式
const (
	LOG_PRO       = "pro"
	LOG_DEBUG     = "debug"
	ENV_PRO   Env = "pro"
	ENV_DEBUG Env = "debug"
	// 秒模版
	DATE_SEC EnvDate = "2006-01-02 15:04:05"
	// 毫秒模版
	DATE_MSEC EnvDate = "2006-01-02 15:04:05.000"
)

var LOC, _ = time.LoadLocation("Local")

// var errorLogger *zap.SugaredLogger
type logsInfo struct {
	m map[string]*zap.SugaredLogger
	sync.RWMutex
	errorTrue bool
}

type log_config struct {
	WithMaxAge       int // 保存多久，单位小时
	WithRotationTime int // 多久切割一次，单位小时
	Env              Env
	formDate         EnvDate // 毫秒或者s 模版 2006-01-02 15:04:05.000
	//one sync.Once
}
type LogOption func(*log_config)

var g_config = &log_config{
	WithMaxAge:       10 * 24,  // 保留日志最大天数
	WithRotationTime: 24,       // 相隔多少个小时切割一次
	Env:              ENV_PRO,  // 正式和测试，默认正式环境
	formDate:         DATE_SEC, // 默认s
}

var info *logsInfo

func init() {
	info = &logsInfo{
		m: make(map[string]*zap.SugaredLogger),
	}
	// 清理没有打开的日志
	clearLog()
}

// 清理过期日志
func clearLog() {
	var dirFile string
	if runtime.GOOS != "windows" {
		dirFile = "./logs/"
	} else {
		dirFile = ".\\logs\\"
	}
	// 先为目录下现有的日志文件启动监听
	initialFiles, err := filepath.Glob(dirFile + "*20*.log")
	if err != nil {
		return
	}
	ruanFileMap := make(map[string]struct{})
	// 需要排查的文件
	excludeRuanFileMap := make(map[string]struct{})
	for _, logFileName := range initialFiles {
		prefixName, ts, err := getLogDate(logFileName)
		if err != nil {
			continue
		}
		now := time.Now() // 确保now使用和ts相同的时区

		// 获取当前时间的年、月、日部分，并将小时、分钟、秒和纳秒设置为0，以便进行日期比较
		midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, LOC)

		if midnight.Sub(*ts) > time.Duration(getConfig().WithMaxAge+24)*time.Hour {
			// 超过
			//t.Log("超过11天")
			// 删除文件
			removeFile(logFileName)
			if prefixName != "" {
				ruanFileName := fmt.Sprintf("%s.log", prefixName)
				if _, ok := ruanFileMap[ruanFileName]; !ok {
					ruanFileMap[ruanFileName] = struct{}{}
				}
			}

		} else {
			// 以内的
			if prefixName != "" {
				excluderuanFileName := fmt.Sprintf("%s.log", prefixName)
				if _, ok := excludeRuanFileMap[excluderuanFileName]; !ok {
					excludeRuanFileMap[excluderuanFileName] = struct{}{}
				}
			}
		}

	}

	// 去掉排除的，然后删掉其他的
	for k := range ruanFileMap {
		_, ok := excludeRuanFileMap[k]
		if ok {
			continue
		}
		//log.Println("删除文件22", k)
		removeFile(k)
	}
}

// 修改getLogDate函数以同时返回文件名前缀和日期
func getLogDate(logFileName string) (prefix string, logDate *time.Time, err error) {
	// 使用正则表达式同时匹配文件名前缀和固定格式的日期：YYYY-MM-DD
	// 分组1匹配前缀，分组2匹配日期
	re := regexp.MustCompile(`^(.*?)(\d{4}-\d{2}-\d{2})\.log$`)

	// 在字符串中查找匹配的日期和前缀
	match := re.FindStringSubmatch(logFileName)
	if match == nil {
		return "", nil, errors.New("no date found in string")
	}
	if len(match) != 3 {
		return "", nil, errors.New("unexpected number of matches")

	}

	// 解析日期
	const layout = "2006-01-02"

	ts, err := time.ParseInLocation(layout, match[2], LOC)

	if err != nil {
		return match[1], &ts, err
	}

	return match[1], &ts, nil
}

func removeFile(file string) error {
	// 删除文件
	err := os.Remove(file)
	if err != nil {
		return err

	}
	return nil
}

func getConfig() *log_config {
	return g_config
}

func WithMaxAge(withMaxAge int) LogOption {
	return func(obj *log_config) {
		obj.WithMaxAge = withMaxAge
	}
}

func WithRotationTime(withRotationTime int) LogOption {
	return func(obj *log_config) {
		obj.WithRotationTime = withRotationTime
	}
}

// 设置时间格式
func WithDate(date EnvDate) LogOption {
	return func(obj *log_config) {
		obj.formDate = date
	}
}

// Deprecated
func SetConfig(withMaxAge, withRotationTime int) {
	g_config.WithRotationTime = withRotationTime
	g_config.WithMaxAge = withMaxAge
}

// Deprecated: use LogConfig
func SetEnv(env string) {
	g_config.Env = Env(env)
}

// 设置环境， pro （pro 默认info级别，debug 默认 debug日志级别）
func SetLog(env Env, options ...LogOption) {
	g_config.Env = env
	for _, option := range options {
		option(g_config)
	}
}

// 指定写入的文件
func F(fileNameArr ...string) *zap.SugaredLogger {
	var fileName string
	if len(fileNameArr) > 0 && fileNameArr[0] != "" {
		fileName = fileNameArr[0]
	} else {
		fileName = "sign"
	}
	m, ok := info.getMap(fileName)
	if !ok || m == nil {
		if len(fileNameArr) > 1 {
			// 函数晚上一级
			m = info.getSetMap(fileName, 1)
		} else {
			m = info.getSetMap(fileName, 0)
		}
	}
	return m

}

// 落盘
func Sync(fileName string) error {
	defer func() {
		if err := recover(); err != nil {
			F().Errorf("错误panic:%s", string(debug.Stack()))
		}
	}()
	if fileName == "" {
		return errors.New("文件错误")
	}
	m, ok := info.getMap(fileName)
	if ok && m != nil {
		return m.Sync()
	}
	return nil

}

// -------------------------  以下内部调用  -----------------------

func (c *logsInfo) getMap(fileName string) (*zap.SugaredLogger, bool) {
	info.RLock()
	defer info.RUnlock()
	m, ok := info.m[fileName]
	return m, ok

}

func (c *logsInfo) getSetMap(fileName string, line uint8) *zap.SugaredLogger {
	info.Lock()
	defer info.Unlock()
	m, ok := info.m[fileName]
	if !ok {
		m = c.getLog(fileName, line)

		info.m[fileName] = m
	}
	return m
}

var errorWriter zapcore.WriteSyncer

func (c *logsInfo) getLog(name string, line uint8) *zap.SugaredLogger {

	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	//encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "line",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		//EncodeCaller:   zapcore.FullCallerEncoder, // 绝对路径
		EncodeCaller: zapcore.ShortCallerEncoder, // 相对路径
	})

	// 实现两个判断日志等级的interface
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if g_config.Env == ENV_PRO {
			return lvl >= zapcore.InfoLevel
		} else {
			return lvl >= zapcore.DebugLevel
		}
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
	var infoWriter zapcore.WriteSyncer
	//if runtime.GOOS != "windows" {
	// 主要错误日志写两份
	infoWriter, _ = getWriteSyncerInfo(fmt.Sprintf("./logs/%s_info.log", name))
	//errorWriter, _ = getWriteSyncerErr(fmt.Sprintf("./logs/%s_error.log", name)) // 这个主要 error错误在info和 error里面都写一份

	if !c.errorTrue {
		errorWriter, _ = getWriteSyncerErr("./logs/sign_error.log") // 这个主要 error错误在info和 error里面都写一份
		c.errorTrue = true
	}
	//} else {
	//	infoWriter, _ = getWriter_v1_win(fmt.Sprintf("./logs/%s_info.log", name))
	//	errorWriter, _ = getWriter_v2_win(fmt.Sprintf("./logs/%s_error.log", name))
	//}

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
	)

	caller := []zap.Option{
		zap.AddCaller(),
	}
	if line > 0 {
		caller = append(caller, zap.AddCallerSkip(1))
	}

	//logs := zap.New(core, zap.AddCaller()) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	errorLogger := zap.New(core, caller...).Sugar() // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑
	//errorLogger.Sync() // 落盘
	return errorLogger

}

// 自定义日志输出时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(string(g_config.formDate)))
}

//git archive --format=zip --prefix=zlog-v0.1.0/ v0.1.0 -o zlog-v0.1.0.zip
