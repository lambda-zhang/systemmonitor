package systemmonitor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Cpustateinfo struct {
	Cpu_idle       uint64 // time spent in the idle task
	Cpu_total      uint64 // total of all time fields
	Cpu_permillage int    //usage, 50 meas 5%

	Avg1min  float32
	Avg5min  float32
	Avg15min float32

	pre_Cpu_idle  uint64
	pre_Cpu_total uint64
	cur_Cpu_idle  uint64
	cur_Cpu_total uint64
}

func (this *Cpustateinfo) getloadavg() error {
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
		err = fmt.Errorf("Avg1min unknown")
		return err
	} else {
		this.Avg1min = float32(val)
	}
	val, err = strconv.ParseFloat(loadavgs[1], 32)
	if err != nil {
		err = fmt.Errorf("Avg5min unknown")
		return err
	} else {
		this.Avg5min = float32(val)
	}
	val, err = strconv.ParseFloat(loadavgs[2], 32)
	if err != nil {
		err = fmt.Errorf("Avg15min unknown")
		return err
	} else {
		this.Avg15min = float32(val)
	}

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

func (this *Cpustateinfo) getcpuusage() error {
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
		this.cur_Cpu_total, this.cur_Cpu_idle, _ = _getcpuusage(fields)
	}
	file.Close()

	if this.cur_Cpu_total > 0 && this.pre_Cpu_total > 0 {
		this.Cpu_total = this.cur_Cpu_total - this.pre_Cpu_total
		this.Cpu_idle = this.cur_Cpu_idle - this.pre_Cpu_idle
	}

	if this.Cpu_total < 1 || this.Cpu_idle < 1 {
		this.Cpu_permillage = 0
	} else if this.Cpu_total < this.Cpu_idle {
		this.Cpu_permillage = 1000
	} else {
		this.Cpu_permillage = int((float64(this.Cpu_total-this.Cpu_idle) / float64(this.Cpu_total)) * 1000)
	}
	if this.Cpu_permillage > 1000 {
		this.Cpu_permillage = 1000
	} else if this.Cpu_permillage < 1 {
		this.Cpu_permillage = 0
	}

	if this.cur_Cpu_total > 0 {
		this.pre_Cpu_idle = this.cur_Cpu_idle
		this.pre_Cpu_total = this.cur_Cpu_total
	}

	return nil
}

func (this *Cpustateinfo) getcpustateinfo() error {
	err := this.getloadavg()
	if err != nil {
		return err
	}
	err = this.getcpuusage()
	if err != nil {
		return err
	}
	return nil
}
