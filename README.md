# zlog

基于 [zap](https://github.com/uber-go/zap) 封装的高性能 Go 日志工具，聚焦于多业务独立落盘、自动切割与错误日志监控，开箱即用。

## 功能特性
- JSON 日志格式，内置调用方信息，可选毫秒级时间戳。
- 通过 `F("service")` 按需创建业务日志，同时自动写入共享的 `sign_error` 错误日志。
- 同时支持全局模式与 `Manager` 实例模式，避免多方依赖互相覆盖配置。
- 使用 `WithMaxAge`、`WithRotationTime`、`WithDate` 等选项快速配置保留时长与切割周期。
- 预设环境映射 zap 等级：`ENV_DEBUG`、`ENV_INFO`、`ENV_WARN`、`ENV_ERROR`、`ENV_DPANIC`、`ENV_PANIC`、`ENV_FATAL` 等；默认仅 `ENV_DEBUG` 同步输出到控制台。
- `SetZapOut` 将标准库 `log` 输出接入 zlog 的滚动日志。
- 集成 `fsnotify` + `tail` 的错误日志监听能力，可通过回调或通道实时处理异常。

## 安装

```bash
go get github.com/Xuzan9396/zlog
```

最低要求 Go 1.16。

## 使用方式

### 全局模式（向下兼容）

原有 API 完全保留，直接调用即可：

```go
package main

import (
	"errors"

	"github.com/Xuzan9396/zlog"
)

func main() {
	zlog.SetLog(
		zlog.ENV_DEBUG,
		zlog.WithMaxAge(10*24),        // 按小时设置保留时长
		zlog.WithRotationTime(24),     // 按小时设置切割间隔
		zlog.WithDate(zlog.DATE_MSEC), // 可选：启用毫秒时间戳
	)

	zlog.F("checkout").Info("checkout started")
	zlog.F("checkout").Error(errors.New("payment failed"))
	zlog.Info("default channel entry")

	// 手动持久化指定日志
	_ = zlog.Sync("checkout")
}
```

### 实例化模式（推荐在多方依赖场景）

当第三方库与主工程同时依赖 zlog 时，可通过 `Manager` 实例隔离配置，避免互相覆盖：

```go
mgr := zlog.NewManager()
mgr.SetLog(
	zlog.ENV_WARN,
	zlog.WithRotationTime(12),
)

svcLog := mgr.F("payment")
svcLog.Warn("order pending")

// 单独刷新指定实例的日志
_ = mgr.Sync("payment")
```

实例同样提供 `SetZapOut`、`SetDebugLevel`、`WithLevel` 等功能，语义与全局函数一致。

## 配置说明
- `SetLog(env Env, options ...LogOption)`: 统一入口设置环境和参数。
- `WithMaxAge(hours int)`: 日志归档保留时长（单位：小时）。
- `WithRotationTime(hours int)`: 日志切割周期（单位：小时）。
- `WithDate(format EnvDate)`: 切换秒级（`DATE_SEC`）或毫秒级（`DATE_MSEC`）时间格式。
- `WithLevel(level zapcore.Level)`: 在保留环境语义的同时强制指定 zap 等级。
- `SetDebugLevel()`: 运行期间动态调整全局日志等级为 Debug。
- `SetZapOut(path string)`: 将标准库 `log` 输出到指定的滚动日志文件。
- `NewManager(options ...LogOption)`: 创建独立实例，API 与全局保持一致（`mgr.F`、`mgr.Sync`、`mgr.SetZapOut` 等）。

### 等级映射参考

| Env        | Zap Level      | 说明                       |
| ---------- | -------------- | -------------------------- |
| `ENV_DEBUG`  | `DebugLevel`    | 同时输出到控制台与文件       |
| `ENV_INFO`   | `InfoLevel`     | 默认生产级别，仅写文件       |
| `ENV_WARN`   | `WarnLevel`     | 告警级别                   |
| `ENV_ERROR`  | `ErrorLevel`    | 错误级别                   |
| `ENV_DPANIC` | `DPanicLevel`   | 开发环境触发 panic          |
| `ENV_PANIC`  | `PanicLevel`    | 写入日志后 panic            |
| `ENV_FATAL`  | `FatalLevel`    | 写入日志后 `os.Exit(1)`     |

可通过 `WithLevel` 或 `Manager.SetLevel` 自定义等级。

## 错误日志监听

```go
package main

import (
	"log"
	"time"

	"github.com/Xuzan9396/zlog"
)

func main() {
	go func() {
		for i := 0; ; i++ {
			time.Sleep(time.Second)
			zlog.F().Errorf("simulated failure %d", i)
		}
	}()

	ch, err := zlog.WatchErr()
	if err != nil {
		log.Fatal(err)
	}

	for msg := range ch {
		log.Printf("error tail: %s\n", msg)
	}
}
```

如果更偏好回调方式，可使用 `WatchErrCallback`。

## 项目结构
- `config.go`: 配置默认值与 Option 定义。
- `manager.go`: 实例化入口与全局兼容 API。
- `registry.go`: Logger 注册与 zap Core 管理。
- `cleanup.go`: 历史日志清理逻辑。
- `environment.go`: 目录、时区与初始化流程。
- `zlog_unix.go` / `zlog_window.go`: 不同系统下的滚动写入实现与 `SetZapOut`。
- `zwatch.go`: 错误日志监听实现。
- `zap.go`、`syslog.go`: 简化调用的包装函数。

结构化拆分后，业务逻辑保持不变，但更易于维护与扩展。

## 版本记录
- `v0.1.3`: 新增 `WithDate` 支持毫秒时间，完善日志清理、错误监听回调。
- `v0.1.1`: 引入 `SetLog` 选项式配置，合并错误日志输出，支持 zap 重定向标准日志。
