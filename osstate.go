package systemmonitor

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// OSinfo 操作系统信息
type OSinfo struct {
	UpTime        int64
	StartTime     int64
	UsePermillage int //usage, 50 meas 5%

	Arch           string
	Os             string
	KernelVersion  string
	KernelHostname string
	NumCPU         int
}

func (oinfo *OSinfo) getcpuinfo() (int, error) {
	lists, err := listdir("/sys/bus/cpu/devices/")
	if err == nil {
		oinfo.NumCPU = len(lists)
		return oinfo.NumCPU, err
	}
	return 0, err
}

func (oinfo *OSinfo) getuptime() error {
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
	oinfo.UpTime = upsec
	oinfo.StartTime = starttime

	val, err = strconv.ParseFloat(uptimes[1], 64)
	if err != nil {
		err = fmt.Errorf("idletime unknown")
		return err
	}

	cpunum, _ := oinfo.getcpuinfo()
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
	oinfo.UsePermillage = UseAge
	return err
}

func (oinfo *OSinfo) getkernelinfo() error {
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
	oinfo.KernelVersion = version[0] + "-" + version[2]
	oinfo.KernelHostname, err = os.Hostname()
	return err
}

func (oinfo *OSinfo) getosinfo() error {
	oinfo.Arch = runtime.GOARCH
	oinfo.Os = runtime.GOOS

	err := oinfo.getkernelinfo()
	if err != nil {
		return err
	}
	_, err = oinfo.getcpuinfo()
	if err != nil {
		return err
	}
	err = oinfo.getuptime()
	if err != nil {
		return err
	}
	return err
}
