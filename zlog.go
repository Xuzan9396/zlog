package zlog

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//var errorLogger *zap.SugaredLogger
type logsInfo struct {
	m map[string]*zap.SugaredLogger
	sync.Mutex
	done uint32
}

var info *logsInfo

func init()  {
	info = &logsInfo{
		m:make(map[string]*zap.SugaredLogger),
	}
}


// 传入的函数f有返回值error，如果初始化失败，需要返回失败的error
// Do方法会把这个error返回给调用者
func (o *logsInfo) do(f func() error) error {
	if atomic.LoadUint32(&o.done) == 1 { //fast path
		return nil
	}
	return o.slowDo(f)
}
// 如果还没有初始化
func (o *logsInfo) slowDo(f func() error) error {
	o.Lock()
	defer o.Unlock()
	var err error
	if o.done == 0 { // 双检查，还没有初始化
		err = f()
		if err == nil { // 初始化成功才将标记置为已初始化
			atomic.StoreUint32(&o.done, 1)
		}
	}
	return err
}

func F(fileName string )  *zap.SugaredLogger{
	if fileName == "" {
		log.Fatal("文件为空!")
	}

	info.do(func() error {
		var m *zap.SugaredLogger
		var ok bool
		m,ok = info.m[fileName]

		if !ok {
			//log.Println("我执行了几次")
			m= initTwo(fileName)
			info.m[fileName] = m
		}
		return nil
	})

	return info.m[fileName]
}

func initTwo(name string ) *zap.SugaredLogger{
	if name == "" {
		log.Fatal("name为空!")
	}

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
	infoWriter,_ := getWriter_v2(fmt.Sprintf("./logs/%s_info.log",name))
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



func getWriter_v2(fileName string ) (zapcore.WriteSyncer, error) {

	fileWriter, err := rotatelogs.New(
		// %Y-%m-%d %H:%M:%S
		//strings.ReplaceAll(fileName,"logs/","logs/history/")
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(10*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
		//rotatelogs.WithMaxAge(20*time.Second),
		//rotatelogs.WithRotationTime(5*time.Second),
	)
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	//return zapcore.AddSync(fileWriter), err
}



//func Debug(args ...interface{}) {
//	errorLogger.Debug(args...)
//}
//
//func Debugf(template string, args ...interface{}) {
//	errorLogger.Debugf(template, args...)
//}
//
//func Info(args ...interface{}) {
//	errorLogger.Info(args...)
//}
//
//func Infof(template string, args ...interface{}) {
//	errorLogger.Infof(template, args...)
//}
//
//func Warn(args ...interface{}) {
//	errorLogger.Warn(args...)
//}
//
//func Warnf(template string, args ...interface{}) {
//	errorLogger.Warnf(template, args...)
//}
//
//func Error(args ...interface{}) {
//	errorLogger.Error(args...)
//}
//
//func Errorf(template string, args ...interface{}) {
//	errorLogger.Errorf(template, args...)
//}
//
//func DPanic(args ...interface{}) {
//	errorLogger.DPanic(args...)
//}
//
//func DPanicf(template string, args ...interface{}) {
//	errorLogger.DPanicf(template, args...)
//}
//
//func Panic(args ...interface{}) {
//	errorLogger.Panic(args...)
//}
//
//func Panicf(template string, args ...interface{}) {
//	errorLogger.Panicf(template, args...)
//}
//
//func Fatal(args ...interface{}) {
//	errorLogger.Fatal(args...)
//}
//
//func Fatalf(template string, args ...interface{}) {
//	errorLogger.Fatalf(template, args...)
//}