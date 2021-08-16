package systemmonitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Cpustateinfo CPU信息
type Cpustateinfo struct {
	CPUIdle       uint64 // time spent in the idle task
	CPUTotal      uint64 // total of all time fields
	CPUPermillage int    //usage, 50 meas 5%

	Avg1min  float32
	Avg5min  float32
	Avg15min float32

	preCPUIdle  uint64
	preCPUTotal uint64
	curCPUIdle  uint64
	curCPUTotal uint64
}

func (cpu *Cpustateinfo) getloadavg() error {
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
	cpu.Avg1min = float32(val)
	val, err = strconv.ParseFloat(loadavgs[1], 32)
	if err != nil {
		err = fmt.Errorf("avg5min unknown")
		return err
	}
	cpu.Avg5min = float32(val)
	val, err = strconv.ParseFloat(loadavgs[2], 32)
	if err != nil {
		err = fmt.Errorf("avg15min unknown")
		return err
	}
	cpu.Avg15min = float32(val)

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

func (cpu *Cpustateinfo) getcpuusage() error {
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
		if len(fields) < 1 || fields[0] != "cpu" {
			continue
		}
		cpu.curCPUTotal, cpu.curCPUIdle, _ = _getcpuusage(fields)
	}
	file.Close()

	if cpu.curCPUTotal > 0 && cpu.preCPUTotal > 0 {
		cpu.CPUTotal = cpu.curCPUTotal - cpu.preCPUTotal
		cpu.CPUIdle = cpu.curCPUIdle - cpu.preCPUIdle
	}

	if cpu.CPUTotal < 1 || cpu.CPUIdle < 1 {
		cpu.CPUPermillage = 0
	} else if cpu.CPUTotal < cpu.CPUIdle {
		cpu.CPUPermillage = 1000
	} else {
		cpu.CPUPermillage = int((float64(cpu.CPUTotal-cpu.CPUIdle) / float64(cpu.CPUTotal)) * 1000)
	}
	if cpu.CPUPermillage > 1000 {
		cpu.CPUPermillage = 1000
	} else if cpu.CPUPermillage < 1 {
		cpu.CPUPermillage = 0
	}

	if cpu.curCPUTotal > 0 {
		cpu.preCPUIdle = cpu.curCPUIdle
		cpu.preCPUTotal = cpu.curCPUTotal
	}

	return nil
}

func (cpu *Cpustateinfo) getcpustateinfo() error {
	err := cpu.getloadavg()
	if err != nil {
		return err
	}
	err = cpu.getcpuusage()
	if err != nil {
		return err
	}
	return nil
}
