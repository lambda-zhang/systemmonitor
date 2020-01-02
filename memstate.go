package systemmonitor

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

// Meminfo 内存信息
type Meminfo struct {
	MemTotal         uint64
	MemAvailable     uint64
	MemUsePermillage int //usage, 50 meas 5%

	SwapTotal         uint64
	SwapFree          uint64
	SwapUsePermillage int //usage, 50 meas 5%
}

func (meminfo *Meminfo) getmeminfo() error {
	file, err := os.Open("/proc/meminfo")
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
		if len(fields) != 3 || fields[0] == "" {
			continue
		}
		val, numerr := strconv.ParseUint(fields[1], 10, 64)
		if numerr != nil {
			continue
		}

		if fields[0] == "MemTotal:" {
			meminfo.MemTotal = val * 1024

		} else if fields[0] == "MemAvailable:" {
			meminfo.MemAvailable = val * 1024

		} else if fields[0] == "SwapTotal:" {
			meminfo.SwapTotal = val * 1024

		} else if fields[0] == "SwapFree:" {
			meminfo.SwapFree = val * 1024

		}
	}
	file.Close()

	if meminfo.MemTotal > 0 {
		meminfo.MemUsePermillage = 1000 - int(float64(meminfo.MemAvailable)/float64(meminfo.MemTotal)*1000)
	} else {
		meminfo.MemUsePermillage = 0
	}
	if meminfo.MemUsePermillage > 1000 {
		meminfo.MemUsePermillage = 1000
	}
	if meminfo.SwapTotal > 0 {
		meminfo.SwapUsePermillage = 1000 - int(float64(meminfo.SwapFree)/float64(meminfo.SwapTotal)*1000)
	} else {
		meminfo.SwapUsePermillage = 0
	}
	if meminfo.SwapUsePermillage > 1000 {
		meminfo.SwapUsePermillage = 1000
	}
	return nil
}
