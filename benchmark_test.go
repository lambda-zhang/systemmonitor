package systemmonitor

import (
	"testing"
)

func Benchmark_main(b *testing.B) {
	callback := func(sysinfo *SysInfo) {}
	var periodSec int = 1
	sm := New(periodSec, callback)
	for i := 0; i < b.N; i++ { //use b.N for looping
		sm.Getsysteminfo()
	}
}
