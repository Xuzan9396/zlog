# zlog

基于 [zap](https://github.com/uber-go/zap) 封装的高性能 Go 日志工具，聚焦于多业务独立落盘、自动切割与错误日志监控，开箱即用。

## 功能特性
- JSON 日志格式，内置调用方信息，可选毫秒级时间戳。
- 通过 `F("service")` 按需创建业务日志，同时自动写入共享的错误日志。
- 同时支持全局模式与 `Manager` 实例模式，避免多方依赖互相覆盖配置。
- 使用 `WithMaxAge`、`WithRotationTime`、`WithDate` 等选项快速配置保留时长与切割周期。
- 预设环境映射 zap 等级：`ENV_DEBUG`、`ENV_INFO`、`ENV_WARN`、`ENV_ERROR`、`ENV_DPANIC`、`ENV_PANIC`、`ENV_FATAL` 等；默认仅 `ENV_DEBUG` 同步输出到控制台。
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

### 动态调整方法
- **级别调整**：
  - `SetDebugLevel()`: 调整为 Debug 级别
  - `SetInfoLevel()`: 调整为 Info 级别
  - `SetWarnLevel()`: 调整为 Warn 级别
  - `SetErrorLevel()`: 调整为 Error 级别
  - `SetDPanicLevel()`: 调整为 DPanic 级别
  - `SetPanicLevel()`: 调整为 Panic 级别
  - `SetFatalLevel()`: 调整为 Fatal 级别
- **输出模式**：
  - `SetConsoleOnly(bool)`: 动态切换仅终端输出模式
- **清理控制**：
  - `CleanupLogs()`: 手动触发日志清理
  - `StopCleanupTask()`: 停止后台清理任务
  - `IsCleanupRunning()`: 查询清理任务状态

### 其他 API
- `SetZapOut(path string)`: 将标准库 `log` 输出到指定的滚动日志文件。
- `NewManager(options ...LogOption)`: 创建独立实例，API 与全局保持一致（支持所有上述方法）。
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
