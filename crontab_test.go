package zlog

import (
	"errors"
	"testing"
)

func Test_log(t *testing.T)  {

	i := 1

	for{
		F("xuzan").Infof("%s","测试下豪不好2")
		F("xuzan2").Error(errors.New("错误"))
		if i >= 1000{
			break
		}
		i++
		//glg.Infof("in main args:%v", os.Args)
		//glg.Errorf("eerror %v", "error")
	}
}
