package zlog

import (
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func TestWathch(t *testing.T) {
	ch, err := WatchErr()
	if err != nil {
		t.Error(err)
		return
	}
	go func() {
		i := 1
		for {
			time.Sleep(time.Second)
			res := atomic.LoadUint32(&curTailNum)
			log.Println("协程数量:", res)
			i++
			F().Errorf("测试下好不好:%d", i)
		}
	}()
	for {
		select {
		case name := <-ch:
			t.Log("文件变化:", name)
		}
	}
}
