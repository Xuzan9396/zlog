package zlog

import (
	"strconv"
	"testing"
	"time"
)

func Test_log(t *testing.T)  {

	i := 1

	SetEnv(LOG_DEBUG)
	for{
		F().Infof("%s","测试下好不好" + strconv.Itoa(i))
		//F("xuzan2").Error(errors.New("错误"))
		if i >= 1000{
			break
		}
		t.Log(i,F() == nil ,F())
		time.Sleep(1*time.Second)

		i++
		//glg.Infof("in main args:%v", os.Args)
		//glg.Errorf("eerror %v", "error")
	}
	time.Sleep(5*time.Second)
}


func Test_logv2(t *testing.T)  {

	i := 1

	SetEnv(LOG_DEBUG)
	for{
		Infof("%s","测试下好不好" + strconv.Itoa(i))
		//F("xuzan2").Error(errors.New("错误"))
		if i >= 1000{
			break
		}
		time.Sleep(1*time.Second)
		i++
		//glg.Infof("in main args:%v", os.Args)
		//glg.Errorf("eerror %v", "error")
	}
	time.Sleep(5*time.Second)
}

