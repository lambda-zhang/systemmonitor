package systemmonitor

import (
	"testing"
)

func Test_main(test *testing.T) {
	callback := func(sysinfo *SysInfo) {}
	var period_sec int = 1
	sm := New(period_sec, callback)
	sm.Getsysteminfo()
}
