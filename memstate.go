package systemmonitor

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

type Meminfo struct {
	MemTotal          uint64
	MemAvailable      uint64
	MemUsePermillage int //usage, 50 meas 5%

	SwapTotal          uint64
	SwapFree           uint64
	SwapUsePermillage int //usage, 50 meas 5%
}

func (this *Meminfo) getmeminfo() error {
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
			this.MemTotal = val * 1024

		} else if fields[0] == "MemAvailable:" {
			this.MemAvailable = val * 1024

		} else if fields[0] == "SwapTotal:" {
			this.SwapTotal = val * 1024

		} else if fields[0] == "SwapFree:" {
			this.SwapFree = val * 1024

		}
	}
	file.Close()

	if this.MemTotal > 0 {
		this.MemUsePermillage= 1000 - int(float64(this.MemAvailable) / float64(this.MemTotal) * 1000)
	} else {
		this.MemUsePermillage = 0
	}
	if this.MemUsePermillage > 1000 {
		this.MemUsePermillage = 1000
	}
	if this.SwapTotal > 0 {
		this.SwapUsePermillage = 1000 - int(float64(this.SwapFree) / float64(this.SwapTotal) * 1000)
	} else {
		this.SwapUsePermillage = 0
	}
	if this.SwapUsePermillage > 1000 {
		this.SwapUsePermillage = 1000
	}
	return nil
}
