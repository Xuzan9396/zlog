# zlog
#### 封装日志,按照文件分类,有软件接和切分文件
下面例子：会在当前项目logs文件下面生成 xuzan_info.log 和 xuzan_error.log 为软连接，具体文件是带天数的时间  如果是调用error级别函数，则在 _error.log 存一份错误，在_info.log存一份，如只调用了info级别则在_info.log 存一份
```go
package main

import (
	"errors"
	"github.com/Xuzan9396/zlog"
)

func main()  {
	zlog.SetEnv(zlog.LOG_DEBUG)
	//zlog.SetConfig(设置保留时间单位小时,设置多久切割一次单位小时) 如果不调用默认 10天 24小时切割一次
	zlog.F("xuzan").Info("111")
	zlog.F("xuzan").Error(errors.New("错误"))
	zlog.F().Error(errors.New("错误"))

}

// 会在logs 文件下面生成 xuzan_info.log 和 xuzan_error.log 为软连接，具体文件是带天数的时间  如果是调用error级别函数，则在 _error.log 存一份错误，在_info.log存一份
// 不填写默认名，默认写入sign_xxx.log 文件

```