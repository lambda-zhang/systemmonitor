package systemmonitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// CPUstateinfo CPU信息
type CPUstateinfo struct {
	CPUIdle       uint64 // time spent in the idle task
	CPUTotal      uint64 // total of all time fields
	CPUPermillage int    //usage, 50 meas 5%

	preCPUIdle  uint64
	preCPUTotal uint64
	curCPUIdle  uint64
	curCPUTotal uint64
}

// CPUstateinfos 多核CPU信息
type CPUstateinfos struct {
	CPUs   map[string]*CPUstateinfo
	CPUNum int

	Avg1min  float32
	Avg5min  float32
	Avg15min float32
}

func (cpus *CPUstateinfos) getloadavg() error {
	loadavg, err := readFile2String("/proc/loadavg")
	if err != nil {
		return err
	}

	loadavgs := strings.Fields(loadavg)
	if len(loadavgs) < 3 {
		err = fmt.Errorf("got loadavg failed")
		return err
	}
	if strings.Count(loadavgs[0], "") < 2 || strings.Count(loadavgs[1], "") < 2 || strings.Count(loadavgs[2], "") < 2 {
		err = fmt.Errorf("loadavg string invalied")
		return err
	}

	var val float64
	val, err = strconv.ParseFloat(loadavgs[0], 32)
	if err != nil {
		err = fmt.Errorf("avg1min unknown")
		return err
	}
	cpus.Avg1min = float32(val)
	val, err = strconv.ParseFloat(loadavgs[1], 32)
	if err != nil {
		err = fmt.Errorf("avg5min unknown")
		return err
	}
	cpus.Avg5min = float32(val)
	val, err = strconv.ParseFloat(loadavgs[2], 32)
	if err != nil {
		err = fmt.Errorf("avg15min unknown")
		return err
	}
	cpus.Avg15min = float32(val)

	return nil
}

func _getcpuusage(fields []string) (uint64, uint64, error) {
	var total uint64
	var idle uint64
	sz := len(fields)
	for i := 1; i < sz; i++ {
		val, err := strconv.ParseUint(fields[i], 10, 64)
		if err != nil {
			continue
		}

		total += val
		if i == 4 {
			idle = val
		}
	}
	return total, idle, nil
}

func (cpus *CPUstateinfos) getcpuusage() error {
	if cpus.CPUs == nil {
		cpus.CPUs = make(map[string]*CPUstateinfo)
		cpus.CPUNum = 0
	}

	file, err := os.Open("/proc/stat")
	if err != nil {
		panic(err)
	}

	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		fields := strings.Fields(string(line))
		if len(fields) < 1 || (!strings.HasPrefix(fields[0], "cpu")) {
			continue
		}
		curCPUTotal, curCPUIdle, _ := _getcpuusage(fields)
		if _, ok := cpus.CPUs[fields[0]]; !ok {
			cpus.CPUs[fields[0]] = &CPUstateinfo{}
			cpus.CPUNum = 0
		}
		curCPU := cpus.CPUs[fields[0]]
		curCPU.curCPUTotal = curCPUTotal
		curCPU.curCPUIdle = curCPUIdle
		cpus.CPUNum = cpus.CPUNum + 1
	}
	file.Close()

	for _, v := range cpus.CPUs {
		if v.curCPUTotal > 0 && v.preCPUTotal > 0 {
			v.CPUTotal = v.curCPUTotal - v.preCPUTotal
			v.CPUIdle = v.curCPUIdle - v.preCPUIdle
		}

		if v.CPUTotal < 1 {
			v.CPUPermillage = 0
		} else if v.CPUTotal < v.CPUIdle {
			v.CPUPermillage = 1000
		} else {
			v.CPUPermillage = int((float64(v.CPUTotal-v.CPUIdle) / float64(v.CPUTotal)) * 1000)
		}
		if v.CPUPermillage > 1000 {
			v.CPUPermillage = 1000
		} else if v.CPUPermillage < 1 {
			v.CPUPermillage = 0
		}

		if v.curCPUTotal > 0 {
			v.preCPUIdle = v.curCPUIdle
			v.preCPUTotal = v.curCPUTotal
		}
	}

	return nil
}

func (cpus *CPUstateinfos) getcpustateinfo() error {
	err := cpus.getloadavg()
	if err != nil {
		return err
	}
	err = cpus.getcpuusage()
	if err != nil {
		return err
	}
	return nil
}
