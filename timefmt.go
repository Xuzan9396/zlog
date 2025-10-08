package zlog

import (
	"time"

	"go.uber.org/zap/zapcore"
)

func newTimeEncoder(provider configProvider) func(time.Time, zapcore.PrimitiveArrayEncoder) {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		cfg := provider()
		enc.AppendString(t.In(location).Format(string(cfg.formDate)))
	}
}
