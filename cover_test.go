package systemmonitor

import (
	"testing"
)

func Test_main(test *testing.T) {
	callback := func(sysinfo *SysInfo) {}
	var periodSec int = 1
	sm := New(periodSec, callback)
	sm.Getsysteminfo()
}
