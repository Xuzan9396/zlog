# zlog,基于zap日志包装，快速使用
### 使用场景，有的时候我们需要不同的业务去设置不同日志文件，方便我们快速定位，需要我们随时使用时启动，不需要太多重复性的配置说明，开箱即用

### zlog,是一个根据官方zap高性能日志再封装的一个日志系统，包含日志自动切割，包含自定义日志文件名，启动不需要提前一大推声明，包含

### 简单的配置
```azure
zlog.SetLog(zlog.ENV_DEBUG, zlog.WithMaxAge(10*24), zlog.WithRotationTime(24))
环境，info日志保存10天，切割间隙24小时切割一次

```
- 日志默认json输出，方便解析
- 简单的正式环境和测试环境，默认正式服环境
- 正式服环境 默认是info等级级别
- 测试服环境 默认是debug等级级别，默认会输出stdout端
- error 日志或者panic日志会生成一个在error 文件里面，同时也会在 info 文件里面生成一份
- 文件自动切分，过期时间设置



#### 封装日志生成说明,按照文件分类,有软件接和切分文件
下面例子：会在当前项目logs文件下面生成 xuzan_info.log 和 xuzan_error.log 为软连接，具体文件是带天数的时间  如果是调用error级别函数，则在 _error.log 存一份错误，在_info.log存一份，如只调用了info级别则在_info.log 存一份

## [0.1.1] - 2023-04-03
### 添加
- zlog.SetLog(zlog.ENV_DEBUG, zlog.WithMaxAge(10*24), zlog.WithRotationTime(24))  环境，存活时间，和切分时间
- zlog.SetZapOut("./logs/sys_log") // 系统日志读取通过zap形式输入到日志和终端，正式环境只会写文件，测试环境会写日志和终端输出, 并且自动切分

### 修改
- 废弃zlog.SetConfig()
- 废弃每个文件生成error多个日志,合并成一个sign_error日志

```go
package main

import (
	"errors"
	"github.com/Xuzan9396/zlog"
)

func main()  {

	//zlog.SetConfig(设置保留时间单位小时,设置多久切割一次单位小时) 如果不调用默认 10天 24小时切割一次
	zlog.SetLog(zlog.ENV_DEBUG, zlog.WithMaxAge(10*24), zlog.WithRotationTime(24))
	zlog.F("xuzan").Info("111")
	zlog.Info("111") // 默认sign 文件
	zlog.F("xuzan").Error(errors.New("错误"))
	zlog.F().Error(errors.New("错误"))

}

// 会在logs 文件下面生成 xuzan_info.log 和 xuzan_error.log 为软连接，具体文件是带天数的时间  如果是调用error级别函数，则在 _error.log 存一份错误，在_info.log存一份
// 不填写默认名，默认写入sign_xxx.log 文件

```
