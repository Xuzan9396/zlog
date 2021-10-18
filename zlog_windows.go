package zlog

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

func getWriter_v2(fileName string ) (zapcore.WriteSyncer, error) {
	fileWriter, err := rotatelogs.New(
		// %Y-%m-%d %H:%M:%S
		strings.Replace(fileName, ".log", "", -1)+"%Y-%m-%d.log", // 没有使用go风格反人类的format格式
		rotatelogs.WithMaxAge(time.Duration(g_config.WithMaxAge)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(g_config.WithRotationTime)*time.Hour),
	)
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(fileWriter)), err
	//return zapcore.AddSync(fileWriter), err
}

