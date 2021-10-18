package zlog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sync"
	"time"
)

const  (
	LOG_PRO = "pro"
	LOG_DEBUG = "debug"
)


//var errorLogger *zap.SugaredLogger
type logsInfo struct {
	m map[string]*zap.SugaredLogger
	sync.RWMutex
}

type log_config struct {
	WithMaxAge int  // 保存多久，单位小时
	WithRotationTime int // 多久切割一次，单位小时
	Env string
}

var g_config = log_config{
	WithMaxAge:10*24,
	WithRotationTime: 24,
	Env:LOG_PRO,
}

var info *logsInfo

func init()  {
	info = &logsInfo{
		m:make(map[string]*zap.SugaredLogger),
	}
}
//  withMaxAge 保留几天
// withRotationTime 多久切割一次
func SetConfig(withMaxAge,withRotationTime int )  {
	g_config.WithRotationTime = withRotationTime
	g_config.WithMaxAge = withMaxAge
}

func SetEnv(env string)  {
	g_config.Env = env
}


func (c *logsInfo)getMap(fileName string ) (*zap.SugaredLogger,bool) {
	info.RLock()
	defer info.RUnlock()
	m,ok := info.m[fileName]
	return m,ok

}

func (c *logsInfo)getSetMap(fileName string ) (*zap.SugaredLogger) {
	info.Lock()
	defer info.Unlock()
	m,ok := info.m[fileName]
	if !ok {
		//log.Println("我执行了几次")
		m= getLog(fileName)

		info.m[fileName] = m
	}
	return m
}


func F(fileNameArr ...string )  *zap.SugaredLogger{
	var fileName string
	if fileNameArr == nil || len(fileNameArr) <= 0 || fileNameArr[0] == "" {
		//log.Fatal("文件为空!")
		fileName = "sign"
	}else{
		fileName =fileNameArr[0]

	}
	m,ok :=  info.getMap(fileName)
	if !ok {
		m = info.getSetMap(fileName)
	}
	return m

}


func getLog(name string ) *zap.SugaredLogger{

	// 设置一些基本日志格式 具体含义还比较好理解，直接看zap源码也不难懂
	//encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
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
		return lvl >= zapcore.InfoLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	// 获取 info、error日志文件的io.Writer 抽象 getWriter() 在下方实现
	infoWriter,_ := getWriter_v1(fmt.Sprintf("./logs/%s_info.log",name))
	errorWriter,_ := getWriter_v2(fmt.Sprintf("./logs/%s_error.log",name))

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWriter), errorLevel),
	)

	logs := zap.New(core, zap.AddCaller()) // 需要传入 zap.AddCaller() 才会显示打日志点的文件名和行数, 有点小坑

	errorLogger := logs.Sugar()
	return errorLogger

}

// 自定义日志输出时间格式
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format( "2006-01-02 15:04:05"))
}



// 传入的函数f有返回值error，如果初始化失败，需要返回失败的error
// Do方法会把这个error返回给调用者
//func (o *logsInfo) do(f func() error) error {
//	if atomic.LoadUint32(&o.done) == 1 { //fast path
//		return nil
//	}
//	return o.slowDo(f)
//}
//// 如果还没有初始化
//func (o *logsInfo) slowDo(f func() error) error {
//	o.Lock()
//	defer o.Unlock()
//	var err error
//	if o.done == 0 { // 双检查，还没有初始化
//		err = f()
//		if err == nil { // 初始化成功才将标记置为已初始化
//			atomic.StoreUint32(&o.done, 1)
//		}
//	}
//	return err
//}




