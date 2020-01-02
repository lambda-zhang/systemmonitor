package systemmonitor

import (
	"fmt"
	"runtime"

	cron "github.com/robfig/cron"
)

// SysInfo 系统监视器的
type SysInfo struct {
	OS      OSinfo
	CPU     Cpustateinfo
	Mem     Meminfo
	Net     NetWorkInfo
	Fs      Fsinfo
	Thermal Thermalinfo

	c *cron.Cron
}

// Getsysteminfo 获得系统资源信息
func (si *SysInfo) Getsysteminfo() {
	si.OS.getosinfo()
	si.CPU.getcpustateinfo()
	si.Mem.getmeminfo()
	si.Net.getnetworkstate()
	si.Fs.getfsstate()
	si.Thermal.getThermal()
}

// New 新新系统监视器
func New(periodSec int, callback func(*SysInfo)) *SysInfo {
	var info *SysInfo = &SysInfo{}
	ostype := runtime.GOOS
	if ostype != "linux" {
		panic("support linux only fornow")
	}
	if periodSec < 1 {
		panic("period must >= 1second")
	}
	info.Getsysteminfo()

	spec := fmt.Sprintf("@every %ds", periodSec)
	info.c = cron.New()
	info.c.AddFunc(spec, func() {
		info.Getsysteminfo()
		callback(info)
	})
	return info
}

// Start 开始系统监视器的定时任务
func (si *SysInfo) Start() {
	si.c.Start()
}

// Stop 停止系统监视器的定时任务
func (si *SysInfo) Stop() {
	si.c.Stop()
}
