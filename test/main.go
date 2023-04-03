package main

import (
	"github.com/Xuzan9396/zlog"
	"log"
)

func main() {
	// 设置输出格式
	// 将日志输出到文件
	_ = zlog.SetZapOut("./logs/sys_log")
	log.Println("ssss2323")

}
