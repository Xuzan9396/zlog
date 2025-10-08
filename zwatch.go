package zlog

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
)

// watch 作为共享管道传递错误日志行。
var watch chan string

// allctx / allcancel 管理当前 tail 协程的生命周期，确保只有最新文件在监听。
var allctx context.Context
var allcancel context.CancelFunc

// curTailNum 记录并发 tail 协程数量，便于测试和调试。
var curTailNum uint32

// WatchErrCallback 监听错误日志并在每条日志到来时触发回调。
func WatchErrCallback(callback func(msg string)) error {
	watchch, err := WatchErr()
	if err != nil {
		return err
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("panic:", string(debug.Stack()))
			}
		}()
		for {
			select {
			case msg := <-watchch:
				callback(msg)
			}
		}
	}()
	return nil

}

// WatchErr 返回一个通道，用于实时消费 sign_error 日志内容。
func WatchErr() (chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	//defer watcher.Close()
	watch = make(chan string, 10)
	//defer close(watch)
	watchdirFile := logDir()
	ensureDir(watchdirFile)
	err = runtailGo(watchdirFile)
	if err != nil {
		return nil, err
	}

	// 协程监听目录变化
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// 检查是否是新增文件 || event.Op&fsnotify.Remove == fsnotify.Remove
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Ext(event.Name) == ".log" && strings.Contains(event.Name, "sign_error20") {
						//fmt.Println("detected file:", event.Name)
						runtailGo(watchdirFile)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)

			}
		}
	}()

	// 添加目录到监听列表
	err = watcher.Add(watchdirFile)
	if err != nil {
		return nil, err
	}
	return watch, nil

}

// runtailGo 为匹配的错误日志文件启动 tail 协程。
func runtailGo(dirFile string) error {

	// 先为目录下现有的日志文件启动监听
	initialFiles, err := filepath.Glob(filepath.Join(dirFile, "sign_error20*.log"))
	if err != nil {
		return err
	}
	now := time.Now()
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).Format("2006-01-02")

	for _, logFileName := range initialFiles {
		_, ts, err := getLogDate(logFileName)
		if err != nil {
			continue
		}
		if allcancel != nil {
			allcancel()
			allcancel = nil
		}
		if ts.Format("2006-01-02") == nowDay {
			allctx, allcancel = context.WithCancel(context.Background())
			go tailLogFile(allctx, logFileName)
		}
	}
	return nil
}

// tailLogFile 启动协程监听指定文件尾部，并写入 watch 通道。
func tailLogFile(ctx context.Context, filepath string) {
	//log.Println("开始监听日志文件:", filepath)
	t, err := tail.TailFile(filepath, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}})
	if err != nil {
		return
	}
	defer t.Stop() // 确保在退出函数时停止tail操作
	atomic.AddUint32(&curTailNum, 1)
	defer atomic.AddUint32(&curTailNum, ^uint32(0))
	for {
		select {
		case <-ctx.Done(): // 检查context是否被取消
			//log.Println("关闭了旧的日志监听", filepath)
			return // 退出函数
		case line := <-t.Lines: // 等待新的日志行
			//log.Println("没有收到:", line.Text)
			watch <- line.Text
		}
	}
}
