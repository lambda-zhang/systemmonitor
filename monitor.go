package systemmonitor

import (
	"fmt"
	"runtime"

	cron "github.com/robfig/cron"
)

type SysInfo struct {
	OS      OSinfo
	CPU     Cpustateinfo
	Mem     Meminfo
	Net     NetWorkInfo
	Fs      Fsinfo
	Thermal Thermalinfo

	c *cron.Cron
}

func (this *SysInfo) Getsysteminfo() {
	this.OS.getosinfo()
	this.CPU.getcpustateinfo()
	this.Mem.getmeminfo()
	this.Net.getnetworkstate()
	this.Fs.getfsstate()
	this.Thermal.GetThermal()
}

func New(period_sec int, callback func(*SysInfo)) *SysInfo {
	var info *SysInfo = &SysInfo{}
	ostype := runtime.GOOS
	if ostype != "linux" {
		panic("support linux only fornow")
		return nil
	}
	if period_sec < 1 {
		panic("period must >= 1second")
		return nil
	}
	info.Getsysteminfo()

	spec := fmt.Sprintf("@every %ds", period_sec)
	info.c = cron.New()
	info.c.AddFunc(spec, func() {
		info.Getsysteminfo()
		callback(info)
	})
	return info
}

func (this *SysInfo) Start() {
	this.c.Start()
}

func (this *SysInfo) Stop() {
	this.c.Stop()
}
