package systemmonitor

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type OSinfo struct {
	UpTime        int64
	StartTime     int64
	UsePermillage int //usage, 50 meas 5%

	Arch           string
	Os             string
	KernelVersion  string
	KernelHostname string
	NumCpu         int
}

func (this *OSinfo) getcpuinfo() (int, error) {
	lists, err := listdir("/sys/bus/cpu/devices/")
	if err == nil {
		this.NumCpu = len(lists)
		return this.NumCpu, err
	}
	return 0, err
}

func (this *OSinfo) getuptime() error {
	uptime, err := readFile2String("/proc/uptime")
	if err != nil {
		return err
	}
	nowts := time.Now().Unix()

	uptimes := strings.Fields(uptime)
	if len(uptimes) != 2 {
		err = fmt.Errorf("got uptime failed")
		return err
	}
	if strings.Count(uptimes[0], "") < 2 || strings.Count(uptimes[1], "") < 2 {
		err = fmt.Errorf("uptime string invalied")
		return err
	}

	var val float64
	val, err = strconv.ParseFloat(uptimes[0], 64)
	if err != nil {
		err = fmt.Errorf("uptime unknown")
		return err
	}
	upsec := int64(val)
	starttime := nowts - upsec
	this.UpTime = upsec
	this.StartTime = starttime

	val, err = strconv.ParseFloat(uptimes[1], 64)
	if err != nil {
		err = fmt.Errorf("idletime unknown")
		return err
	}

	cpunum, _ := this.getcpuinfo()
	if cpunum == 0 {
		cpunum = 1
	}

	idlesec := int64(val / float64(cpunum))
	var UseAge int
	if upsec == 0 || idlesec >= upsec {
		UseAge = 0
	} else {
		UseAge = 1000 - int(float64(idlesec)/float64(upsec)*1000)
	}
	this.UsePermillage = UseAge
	return err
}

func (this *OSinfo) getkernelinfo() error {
	kversion, err := readFile2String("/proc/version")
	if err != nil {
		return err
	}
	version := strings.Fields(kversion)
	if len(version) <= 2 {
		err = fmt.Errorf("got kernel version failed")
		return err
	}
	if version[0] != "Linux" || version[1] != "version" || strings.Count(version[2], "") < 2 {
		err = fmt.Errorf("kernel version string invalied")
		return err
	}
	this.KernelVersion = version[0] + "-" + version[2]
	this.KernelHostname, err = os.Hostname()
	return nil
}

func (this *OSinfo) getosinfo() error {
	this.Arch = runtime.GOARCH
	this.Os = runtime.GOOS

	err := this.getkernelinfo()
	if err != nil {
		return err
	}
	_, err = this.getcpuinfo()
	if err != nil {
		return err
	}
	err = this.getuptime()
	if err != nil {
		return err
	}
	return err
}
