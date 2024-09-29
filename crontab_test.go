package zlog

import (
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"testing"
	"time"
)

func Test_log(t *testing.T) {

	i := 1

	//SetEnv(LOG_PRO)
	for {
		F().Errorf("%s", "测试下好不好"+strconv.Itoa(i))
		//F("xuzan2").Error(errors.New("错误"))
		if i >= 1000 {
			break
		}

		i++
	}
	time.Sleep(5 * time.Second)
}

func Test_logv2(t *testing.T) {

	i := 1

	SetEnv(LOG_DEBUG)
	for {
		Infof("%s", "测试下好不好"+strconv.Itoa(i))
		//F("xuzan2").Error(errors.New("错误"))
		if i >= 1000 {
			break
		}
		time.Sleep(1 * time.Second)
		i++
		//glg.Infof("in main args:%v", os.Args)
		//glg.Errorf("eerror %v", "error")
	}
	time.Sleep(5 * time.Second)
}

// 4.3更新
func Test_logv3(t *testing.T) {

	i := 1
	//SetEnv(LOG_PRO)
	SetLog(ENV_DEBUG, WithMaxAge(10*24), WithRotationTime(24), WithDate(DATE_MSEC))
	for {
		F().Infof("%s", "测试下好不好"+strconv.Itoa(i))
		F().Debugf("%s", "debug测试下好不好"+strconv.Itoa(i))
		F().Errorf("%s", "error测试下好不好"+strconv.Itoa(i))
		if i >= 1000 {
			break
		}
		t.Log(i, F() == nil, F())
		time.Sleep(1 * time.Second)

		i++

	}
	time.Sleep(5 * time.Second)
}

func Test_logv4(t *testing.T) {

	i := 1
	//SetEnv(ENV_DEBUG)
	SetLog(ENV_DEBUG)
	//SetLog(ENV_DEBUG, WithMaxAge(10*24), WithRotationTime(24))

	for is := 0; is <= 10; is++ {
		go func(is int) {
			for {

				F(fmt.Sprintf("test_%d", is)).Errorf("%s", fmt.Sprintf("v4_%d_error测试"+strconv.Itoa(i), is))

				//t.Log(i, F() == nil, F())
				time.Sleep(3 * time.Second)

			}
		}(is)
	}

	select {}
}

func Test_logv5(t *testing.T) {

	i := 1
	//SetEnv(ENV_DEBUG)
	SetLog(ENV_DEBUG)
	//SetLog(ENV_DEBUG, WithMaxAge(10*24), WithRotationTime(24))

	for is := 0; is <= 10; is++ {
		go func(is int) {
			for {

				F("test").Errorf("%s", fmt.Sprintf("v5_%d_error测试"+strconv.Itoa(i), is))

				time.Sleep(10 * time.Second)

			}
		}(is)
	}

	select {}
}

func Test_logv6(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r, string(debug.Stack()))
		}
	}()

	// 将日志输出到文件
	_ = SetZapOut("./logs/sys_log")
	go func() {
		log.Printf("%s", "测试下sss")
	}()
	log.Println("set zap test")
	panic("232323")
	<-time.After(2 * time.Second)
}
