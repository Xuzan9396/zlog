//go:build !windows
// +build !windows

package zlog

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"strings"
	"time"
)

func getWriteSyncerInfo(fileName string) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		// %Y-%m-%d %H:%M:%S
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithMaxAge(time.Duration(g_config.WithMaxAge)*time.Hour),             // 保存最大的时间
		rotatelogs.WithRotationTime(time.Duration(g_config.WithRotationTime)*time.Hour), // 切割时间
	)
	if getConfig().Env == "pro" {
		// 只写入文件
		return zapcore.AddSync(fileWriter), err
	} else {
		// 测试环境，则终端和文件都写入
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	}
}

func getWriteSyncerErr(fileName string) (zapcore.WriteSyncer, error) {

	fileWriter, err := rotatelogs.New(
		// %Y-%m-%d %H:%M:%S
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(fileName),
		// gxz 后面开启
		rotatelogs.WithMaxAge(time.Duration(g_config.WithMaxAge)*time.Hour),             // 保存最大的时间
		rotatelogs.WithRotationTime(time.Duration(g_config.WithRotationTime)*time.Hour), // 切割时间
		//
		//rotatelogs.WithMaxAge(time.Duration(getConfig().WithMaxAge)*time.Hour),
		//rotatelogs.WithRotationCount(7),           // 做多保存多少分
		//rotatelogs.WithRotationSize(1024*1024), // 10MB切割 , WithRotationSize  和 WithRotationTime 互斥
		//rotatelogs.WithRotationTime(time.Duration(getConfig().WithRotationTime)*time.Hour),
	)
	return zapcore.AddSync(fileWriter), err
}

// 系统日志log 通过zap写入文件
func SetZapOut(fileName string) error {

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	fileWriter, err := rotatelogs.New(
		// %Y-%m-%d %H:%M:%S
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(fileName),
		rotatelogs.WithRotationCount(7),           // 做多保存多少分
		rotatelogs.WithRotationSize(1024*1024*10), // 10MB切割 , WithRotationSize  和 WithRotationTime 互斥
	)
	// 测试环境，则终端和文件都写入
	var w zapcore.WriteSyncer
	if getConfig().Env == "pro" {
		// 只写入文件
		w = zapcore.AddSync(fileWriter)

	} else {
		// 测试环境，则终端和文件都写入
		w, err = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
		if err != nil {
			return err
		}
	}
	log.SetOutput(w)
	return nil

}
