# zlog
封装日志,按照文件分类
```go
package main

import (
	"errors"
	"github.com/Xuzan9396/zlog"
)

func main()  {
	zlog.
	zlog.F("xuzan").Info("111")
	zlog.F("xuzan").Error(errors.New("错误"))
	zlog.F("xuzan2").Info("111")
	zlog.F("xuzan2").Error(errors.New("错误"))
}


```