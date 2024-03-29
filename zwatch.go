package zlog

// 监听发送错误日志
import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/nxadm/tail"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"time"
)

var watch chan string
var allctx context.Context
var allcancel context.CancelFunc
var curTailNum uint32

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

func WatchErr() (chan string, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	//defer watcher.Close()
	watch = make(chan string, 10)
	//defer close(watch)
	watchdirFile := dirpath
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

func runtailGo(dirFile string) error {

	// 先为目录下现有的日志文件启动监听
	initialFiles, err := filepath.Glob(dirFile + "sign_error20*.log")
	if err != nil {
		return err
	}
	now := time.Now()
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, LOC).Format("2006-01-02")

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

// tailLogFile 用于启动一个新的协程来监听指定的日志文件
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
func ensureDir(dirName string) error {
	// 尝试获取目录的状态，判断目录是否存在
	infos, err := os.Stat(dirName)

	// 如果因为目录不存在而报错，则创建目录
	if os.IsNotExist(err) {
		// 使用MkdirAll而不是Mkdir，以确保创建所有必要的父目录
		return os.MkdirAll(dirName, 0755) // 使用适当的权限
	}

	// 如果有其他错误，返回错误
	if err != nil {
		return err
	}

	// 确保dirName确实是一个目录
	if !infos.IsDir() {
		return os.ErrExist // 或者你可以返回一个更具体的错误
	}

	// 目录已存在，无需创建，返回nil表示成功
	return nil
}
