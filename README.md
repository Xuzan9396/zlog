# zlog

基于 [zap](https://github.com/uber-go/zap) 封装的高性能 Go 日志工具，聚焦于多业务独立落盘、自动切割与错误日志监控，开箱即用。

## 功能特性
- JSON 日志格式，内置调用方信息，可选毫秒级时间戳。
- 通过 `F("service")` 按需创建业务日志，同时自动写入共享的错误日志。
- 同时支持全局模式与 `Manager` 实例模式，避免多方依赖互相覆盖配置。
- 使用 `WithMaxAge`、`WithRotationTime`、`WithDate` 等选项快速配置保留时长与切割周期。
- 预设环境映射 zap 等级：`ENV_DEBUG`、`ENV_INFO`、`ENV_WARN`、`ENV_ERROR`、`ENV_DPANIC`、`ENV_PANIC`、`ENV_FATAL` 等；当 `Env=ENV_DEBUG` 或 `Level=DebugLevel` 时会同步输出到控制台，其他等级仅写文件，除非通过 `SetConsoleOnly(true)` 显式开启终端输出。
- **支持动态修改日志级别**：运行时可通过 `SetDebugLevel()` 或 `SetLevel()` 动态调整日志级别和终端输出，无需重启进程，便于线上问题排查。
- **支持仅终端输出模式**：通过 `SetConsoleOnly(true)` 或 `WithConsoleOnly(true)` 可动态切换为仅输出到终端，不写入文件，适合开发调试场景。
- **智能 f 字段**：
  - **终端输出**：固定显示 `f` 字段，默认 logger 显示 `f=log`，指定名称 logger 显示对应名称（如 `f=api`）
  - **文件输出**：不包含 `f` 字段，因为文件名已标识 logger（节省存储空间）
- **自动定位工作目录**：日志文件统一写入程序运行时工作目录的 `logs` 子目录，避免在各个子目录创建多个 logs 文件夹。
- **后台自动清理**：默认每 24 小时自动清理过期日志，包括不再使用的 logger 的旧日志文件。
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

### 顶层快捷函数

`zap.go` 提供了一组无需先调 `F()` 即可直接打日志的顶层函数，**内部固定使用名为 `zlog` 的通道**（即 `F("zlog", "")`），所以日志会落在 `logs/zlog_info.log`（错误级别还会汇入共享的 error 文件），与默认 logger（前缀来自 `ZLOG_FILE_PREFIX` 或 `log`）**不是同一个文件**。

```go
package main

import (
	"github.com/Xuzan9396/zlog"
)

func main() {
	zlog.SetLog(zlog.ENV_DEBUG)

	zlog.Info("server started")             // -> logs/zlog_info.log
	zlog.Errorf("query failed: %v", "EOF") // -> logs/zlog_info.log + 共享 error 文件
	zlog.Debug("debug payload")
	zlog.Infof("user=%s action=%s", "u1", "login")

	// 完整列表: Info/Error/Debug/Warn/Panic/Fatal 与 Infof/Warnf/Errorf/Panicf/Fatalf
}
```

> 提示：`zap.go` 中 `Warn`/`Warnf` 当前实现存在内部调用偏差（`Warn` 内部调用了 `.Debug()`），如需精确控制告警级别，建议使用 `zlog.F("xxx").Warn(...)` 或全局 `zlog.SetWarnLevel()` 后再写日志。

### 自定义调用栈深度

`F(name string, opt ...string)` 第二个参数若**非空**，会让底层 zap 多跳一层 `AddCallerSkip(1)`（参见 `zlog.go:7-9` 与 `manager.go:107-110`）。当业务在 zlog 之上又封装了一层 wrapper 时，用这个开关可以让日志的 `line` 字段指向**真实业务代码**，而不是 wrapper 内部。

```go
package log

import "github.com/Xuzan9396/zlog"

// 业务自己的薄封装层，希望调用方代码出现在 line 字段
func Info(msg string) {
	zlog.F("biz", "skip").Info(msg) // 第二参数非空 -> 调用栈再跳一层
}
```

不需要调整时直接 `zlog.F("biz")` 即可，第二参数省略不会引入额外开销。

### 动态修改日志级别

在生产环境中，如需临时开启 Debug 日志排查问题，可以动态调整日志级别，无需重启进程：

```go
package main

import (
	"time"
	"github.com/Xuzan9396/zlog"
)

func main() {
	// 初始化为生产环境（不输出到终端）
	zlog.SetLog(zlog.ENV_PRO)

	zlog.F().Info("生产环境 Info 日志（仅写文件）")
	zlog.F().Debug("生产环境 Debug 日志（不会记录）")

	// 动态切换到 Debug 级别（会输出到终端）
	zlog.SetDebugLevel()

	zlog.F().Info("切换后 Info 日志（输出到终端和文件）")
	zlog.F().Debug("切换后 Debug 日志（输出到终端和文件）")

	time.Sleep(1 * time.Second)
}
```

**特性说明：**
- `SetDebugLevel()` 会动态切换日志级别为 Debug，并自动启用终端输出
- 支持所有日志级别的快捷设置方法：`SetDebugLevel()`, `SetInfoLevel()`, `SetWarnLevel()`, `SetErrorLevel()`, `SetDPanicLevel()`, `SetPanicLevel()`, `SetFatalLevel()`
- 若需要任意 `zapcore.Level` 值，全局可使用 `zlog.SetLog(env, zlog.WithLevel(level))` 一次性设定；持有 `*Manager` 实例时则用 `mgr.SetLevel(level)`。`SetLevel` 与 `WithLevel` 都会标记 `levelOverride`，避免后续 `SetLog` 用环境默认等级覆盖（参见 `manager.go:86-97` 与 `config.go:121-127`）
- 已创建的 logger 会自动重建，应用新的配置
- 支持全局模式和 Manager 实例模式
- 线程安全，支持高并发场景（已通过 10万+ 次并发测试和 race detector 检测）

### 仅终端输出模式

适合开发调试场景，仅输出到终端，不写入文件：

```go
package main

import (
	"github.com/Xuzan9396/zlog"
)

func main() {
	// 方式1: 动态设置
	zlog.SetLog(zlog.ENV_DEBUG)
	zlog.SetConsoleOnly(true)

	// 方式2: 使用选项
	zlog.SetLog(zlog.ENV_DEBUG, zlog.WithConsoleOnly(true))

	// 默认 logger（f=log）
	zlog.F().Info("默认日志")
	// 终端输出: {"level":"info","time":"2025-11-18 12:00:00","f":"log","line":"main.go:15","message":"默认日志"}

	// 指定名称的 logger（f=对应名称）
	zlog.F("v1").Info("v1 API 日志")
	// 终端输出: {"level":"info","time":"2025-11-18 12:00:00","f":"v1","line":"main.go:18","message":"v1 API 日志"}

	zlog.F("payment").Warn("支付服务警告")
	// 终端输出: {"level":"warn","time":"2025-11-18 12:00:00","f":"payment","line":"main.go:21","message":"支付服务警告"}

	// 动态切换回文件模式
	zlog.SetConsoleOnly(false)
}
```

**特性说明：**
- 仅终端模式下，日志只输出到 stdout，不写入任何文件
- **终端输出固定包含 `f` 字段**：默认 logger 显示 `f=log`，指定名称 logger 显示对应名称
- **文件输出不包含 `f` 字段**：因为文件名已标识 logger，避免冗余存储
- 可通过 `SetConsoleOnly(bool)` 动态切换模式
- 支持全局模式和 Manager 实例模式

### 后台自动清理

默认启用后台清理任务，每 24 小时自动清理过期日志（包括不再使用的 logger 的旧日志）：

```go
package main

import (
	"time"
	"github.com/Xuzan9396/zlog"
)

func main() {
	// 默认配置：自动清理已启用
	zlog.SetLog(zlog.ENV_INFO)
	// 后台会每 24 小时清理超过 10 天（240小时）的日志

	// 自定义清理配置
	zlog.SetLog(zlog.ENV_INFO,
		zlog.WithMaxAge(7*24),                  // 保留 7 天
		zlog.WithCleanupInterval(12*time.Hour), // 每 12 小时清理一次
	)

	// 禁用自动清理
	zlog.SetLog(zlog.ENV_INFO,
		zlog.WithAutoCleanup(false),
	)

	// 手动触发清理
	zlog.CleanupLogs()

	// 停止后台清理任务
	zlog.StopCleanupTask()

	// 程序退出时停止清理任务（推荐）
	defer zlog.StopCleanupTask()
}
```

**特性说明：**
- 默认每 24 小时自动清理一次，保留 10 天日志
- 扫描整个日志目录，清理所有过期文件（包括不再使用的 logger）
- 自动清理无效的软链接
- 不在程序启动时清理，避免阻塞启动
- 支持通过 `WithAutoCleanup(false)` 禁用自动清理
- 提供 `CleanupLogs()` 手动触发清理

## 配置说明

### 配置选项
- `SetLog(env Env, options ...LogOption)`: 统一入口设置环境和参数。
- `WithMaxAge(hours int)`: 日志归档保留时长（单位：小时），默认 `10*24=240` 小时（10天）。
- `WithRotationTime(hours int)`: 日志切割周期（单位：小时），默认 `24` 小时。
- `WithDate(format EnvDate)`: 切换秒级（`DATE_SEC`）或毫秒级（`DATE_MSEC`）时间格式。
- `WithLevel(level zapcore.Level)`: 在保留环境语义的同时强制指定 zap 等级。
- `WithConsoleOnly(bool)`: 设置为仅终端输出模式（true）或文件模式（false）。
- `WithAutoCleanup(bool)`: 是否启用后台自动清理，默认 `true`。
- `WithCleanupInterval(duration)`: 后台清理间隔，默认 `24 * time.Hour`。
- `WithDefaultName(name string)`: 修改默认 logger 前缀（业务日志写入 `logs/<name>_info.log`），未单独设置错误前缀时会自动派生 `<name>_error` 作为错误日志前缀（参见 `config.go:130-140`）。
- `WithErrorName(name string)`: 单独指定错误日志聚合前缀，覆盖 `WithDefaultName` 的派生规则；所有 logger 的 error 级别都会汇聚到此前缀对应的文件（参见 `config.go:143-150`）。

> **默认 logger 前缀解析顺序**：`WithDefaultName` > 环境变量 `ZLOG_FILE_PREFIX` > `"log"`。错误日志前缀默认为 `<defaultName>_error`，可通过 `WithErrorName` 覆盖。

### 动态调整方法
- **级别调整**：
  - `SetDebugLevel()`: 调整为 Debug 级别
  - `SetInfoLevel()`: 调整为 Info 级别
  - `SetWarnLevel()`: 调整为 Warn 级别
  - `SetErrorLevel()`: 调整为 Error 级别
  - `SetDPanicLevel()`: 调整为 DPanic 级别
  - `SetPanicLevel()`: 调整为 Panic 级别
  - `SetFatalLevel()`: 调整为 Fatal 级别
  - `Manager.SetLevel(level zapcore.Level)`: 实例方法，设置任意 zap 等级（如 `mgr.SetLevel(zapcore.WarnLevel)`），全局未提供同名快捷函数，可通过 `SetLog(env, WithLevel(level))` 达到同等效果
- **输出模式**：
  - `SetConsoleOnly(bool)`: 动态切换仅终端输出模式
- **清理控制**：
  - `CleanupLogs()`: 手动触发日志清理
  - `StopCleanupTask()`: 停止后台清理任务
  - `IsCleanupRunning()`: 查询清理任务状态

### 其他 API
- `SetZapOut(path string)`: 将标准库 `log` 输出到滚动日志文件。**注意**：此入口与主日志切割策略不同，按 `WithRotationCount(7) + WithRotationSize(10MB)` 切割（即最多保留 7 个文件，单文件超 10MB 触发滚动），并非按 `WithRotationTime` 时间切割（参见 `zlog_unix.go:64-65`）。
- `NewManager(options ...LogOption)`: 创建独立实例，API 与全局保持一致（支持所有上述方法）。
- `SetEnv(env string)`: 兼容旧入口，等价于 `SetLog(Env(env))`，仅切换环境（参见 `manager.go:246-248`）。
- `SetConfig(maxAge, rotationTime int)`: 兼容旧入口，仅调整日志保留与切割周期，不重置环境（参见 `manager.go:251-253`）。
- `Manager.UpdateRetention(maxAge, rotationTime int)`: 上面 `SetConfig` 的实例版（参见 `manager.go:76-83`）。
- `Manager.SetLevel(level zapcore.Level)`: 实例直接设置任意 zap 等级，并标记 `levelOverride`（参见 `manager.go:86-97`）。
- 环境变量 `ZLOG_FILE_PREFIX`：设置默认日志前缀（默认 `log`），错误日志会自动追加 `_error`。

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

> **终端输出触发条件**：当 `cfg.Env == ENV_DEBUG` 或 `cfg.Level == zapcore.DebugLevel` 时（参见 `registry.go:109`），日志会在写文件的同时输出到 stdout。`SetDebugLevel()`、`WithLevel(zapcore.DebugLevel)` 都会触发该条件。若想跳过文件直接写终端，请用 `SetConsoleOnly(true)` 或 `WithConsoleOnly(true)`。

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

如果更偏好回调方式，可使用 `WatchErrCallback`，无需自己起 `for range` 协程：

```go
package main

import (
	"log"
	"time"

	"github.com/Xuzan9396/zlog"
)

func main() {
	zlog.SetLog(zlog.ENV_INFO)

	err := zlog.WatchErrCallback(func(msg string) {
		// 此回调由内部 goroutine 触发；切勿在此执行阻塞 I/O，
		// 否则会拖慢错误日志的实时消费（参见 zwatch.go:28-48）。
		log.Printf("error tail: %s\n", msg)
	})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; ; i++ {
		time.Sleep(time.Second)
		zlog.F().Errorf("simulated failure %d", i)
	}
}
```

> **共享通道注意**：`WatchErr` 和 `WatchErrCallback` 内部共用同一个 `watch chan string`（容量 10）。同一进程内只应注册一份消费者，重复注册会争抢同一通道导致回调丢消息。

## 项目结构
- `config.go`: 配置默认值与 Option 定义。
- `manager.go`: 实例化入口与全局兼容 API。
- `registry.go`: Logger 注册与 zap Core 管理。
- `cleanup.go`: 历史日志清理逻辑。
- `environment.go`: 目录、时区与初始化流程。
- `zlog_unix.go` / `zlog_window.go`: 不同系统下的滚动写入实现与 `SetZapOut`。
- `zwatch.go`: 错误日志监听实现。
- `zap.go`: 顶层快捷函数（`Info`/`Error`/`Debug` 等），固定使用 `zlog` 通道写文件。
- `syslog.go`: `CustomLogger` —— 基于标准库 `log` 的独立 stderr 日志包装器，与 zap 无关，可单独使用。
- `timefmt.go`: 共享时间编码器，根据 `Config.formDate`（`DATE_SEC` / `DATE_MSEC`）输出秒级或毫秒级时间戳。
- `zlog.go`: 包对外 API（`F` / `Sync` / `Set*Level` / `SetConsoleOnly`）。

结构化拆分后，业务逻辑保持不变，但更易于维护与扩展。

### 日志文件命名与切割策略

| 类型 | 文件路径 | 切割策略 |
|------|----------|----------|
| 业务日志 | `logs/<name>_info<YYYY-MM-DD>.log` + 软链 `logs/<name>_info.log` | 按时间，由 `WithRotationTime(hours)` 控制（默认 24h） |
| 错误日志（共享） | `logs/<errorName><YYYY-MM-DD>.log` + 软链 `logs/<errorName>.log` | 同上；所有 logger 的 error 级别都汇入此文件（`registry.go:130-141`） |
| `SetZapOut` 重定向 | `logs/<path><YYYY-MM-DD>.log` + 软链 | **特殊**：按 `WithRotationCount(7) + WithRotationSize(10MB)` 切割，与主日志策略不同（`zlog_unix.go:64-65`） |

- `<name>` 由 `F("name")` 决定；`F()` 默认 logger 的 `<name>` 来自 `WithDefaultName` > 环境变量 `ZLOG_FILE_PREFIX` > `"log"`
- `<errorName>` 由 `WithErrorName` 决定，未设置时默认派生为 `<name>_error`
- 顶层 `zlog.Info/Error/...` 使用固定通道 `zlog`，所以会落在 `logs/zlog_info.log`

## 版本记录
- `v1.0.6`: 文档同步（仅 README，无代码改动）
  - 补全顶层快捷函数 `Info` / `Error` / `Debug` / `Warn` / `Panic` / `Fatal` 及对应 `*f` 形式的说明
  - 补全 `SetEnv` / `SetConfig` 兼容入口与 `Manager.UpdateRetention` / `Manager.SetLevel` 实例方法
  - 补全 `WithDefaultName` / `WithErrorName` 配置选项，并明确默认前缀解析顺序
  - 补全 `WatchErrCallback` 完整示例 + 共享通道注意事项
  - 补充 `F` 第二参数（调用栈深度调整）的使用方式
  - 修正终端输出触发条件描述（`Env=ENV_DEBUG` 或 `Level=DebugLevel`）
  - 修正 `SetZapOut` 切割策略（按数量+大小，与主日志按时间不同）与 `syslog.go` 用途描述
  - 新增「日志文件命名与切割策略」小节
- `v1.0.5`: 重大优化版本
  - **优化 f 字段显示策略**：
    - 终端输出固定显示 `f` 字段，默认 logger 显示 `f=log`，便于区分日志来源
    - 文件输出不包含 `f` 字段，因为文件名已标识 logger，节省存储空间
  - **后台自动清理机制**：
    - 默认启用后台清理任务，每 24 小时自动清理过期日志
    - 解决 rotatelogs 局限：清理所有过期日志，包括不再使用的 logger 的旧文件
    - 移除启动时清理，避免阻塞程序启动
    - 新增 `CleanupLogs()`、`StopCleanupTask()`、`IsCleanupRunning()` API
    - 新增 `WithAutoCleanup(bool)` 和 `WithCleanupInterval(duration)` 配置选项
  - **日志目录优化**：
    - 使用程序运行时工作目录的绝对路径，避免在子目录创建多个 logs 文件夹
- `v1.0.4`: 新增重要功能
  - 添加仅终端输出模式：`SetConsoleOnly(bool)` 和 `WithConsoleOnly(bool)`，支持动态切换
  - 智能 f 字段：使用 `F("service_name")` 时自动添加 `f` 字段标识服务，默认 `F()` 不显示该字段
  - 修复日志目录问题：统一使用程序运行时工作目录的 logs 子目录，避免在各个目录创建多个 logs 文件夹
- `v1.0.3`: 添加所有日志级别的快捷设置方法（SetInfoLevel, SetWarnLevel, SetErrorLevel, SetDPanicLevel, SetPanicLevel, SetFatalLevel），完善动态级别切换能力。
- `v1.0.2`: 修复动态修改日志级别功能，支持 PRO 环境下通过 `SetDebugLevel()` 动态启用终端输出；新增并发压测和性能基准测试。
- `v1.0.1`: 稳定版本发布。
- `v1.0.0`: 重构代码结构，提升可维护性。
- `v0.1.3`: 新增 `WithDate` 支持毫秒时间，完善日志清理、错误监听回调。
- `v0.1.1`: 引入 `SetLog` 选项式配置，合并错误日志输出，支持 zap 重定向标准日志。
