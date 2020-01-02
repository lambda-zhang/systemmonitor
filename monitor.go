package systemmonitor

import (
	"fmt"
	"runtime"

	cron "github.com/robfig/cron"
)

// SysInfo 系统监视器的
type SysInfo struct {
	OSEn bool
	OS   OSinfo

	CPUEn bool
	CPU   Cpustateinfo

	MemEn bool
	Mem   Meminfo

	NetEn bool
	Net   NetWorkInfo

	FsEn bool
	Fs   Fsinfo

	ThermalEn bool
	Thermal   Thermalinfo

	PeriodSec int //　任务执行周期，单位是秒

	c *cron.Cron
}

// Getsysteminfo 获得系统资源信息
func (si *SysInfo) Getsysteminfo() {
	if si.OSEn {
		si.OS.getosinfo()
	}
	if si.CPUEn {
		si.CPU.getcpustateinfo()
	}
	if si.MemEn {
		si.Mem.getmeminfo()
	}
	if si.NetEn {
		si.Net.getnetworkstate()
	}
	if si.FsEn {
		si.Fs.getfsstate()
	}
	if si.ThermalEn {
		si.Thermal.getThermal()
	}
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
	info.PeriodSec = periodSec
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
