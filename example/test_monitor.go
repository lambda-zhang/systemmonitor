package main

import (
	"fmt"
	"time"

	s "github.com/lambda-zhang/systemmonitor"
)

func callback(sysinfo *s.SysInfo) {
	fmt.Println("###############################################")
	fmt.Println("os:")
	o := sysinfo.OS
	fmt.Printf("\tStartTime=%s Arch=%s Os=%s KernelVersion=%s KernelHostname=%s NumCpu=%d UpTime=%d UsePermillage=%d‰\n",
		time.Unix(o.StartTime, 0), o.Arch, o.Os, o.KernelVersion, o.KernelHostname, o.NumCPU, o.UpTime, o.UsePermillage)

	fmt.Println("cpu:")
	c := sysinfo.CPU
	for k, v := range c.CPUs {
		fmt.Printf("\t%s Cpu_permillage=%d‰  Avg1min=%f\n", k, v.CPUPermillage, c.Avg1min)
	}

	fmt.Println("memory:")
	m := sysinfo.Mem
	fmt.Printf("\tMemTotal=%d  MemUsePermillage=%d‰  SwapTotal=%d  SwapUsePermillage=%d‰\n", m.MemTotal, m.MemUsePermillage, m.SwapTotal, m.SwapUsePermillage)

	fmt.Println("net:")
	n := sysinfo.Net
	for k, v := range n.Cards {
		fmt.Printf("\t%s: inKBps=%d outKBps=%d\n", k, v.InBytes/uint64(sysinfo.PeriodSec)/1024, v.OutBytes/uint64(sysinfo.PeriodSec)/1024)
	}
	fmt.Printf("\ttotal TcpConnections=%d ESTABLISHED=%d TCP_LISTEN=%d\n",
		n.TCP.TCPConnections+n.TCP6.TCPConnections, n.TCP.TCPEstablished+n.TCP6.TCPEstablished, n.TCP.TCPListen+n.TCP6.TCPListen)

	fmt.Println("disk:")
	f := sysinfo.Fs
	for k, v := range f.Disks {
		fmt.Printf("\t%s FsVfstype=%s BytesAll=%d BytesUsedPermillage=%d‰ ReadBytes=%dKBps  WriteBytes=%dKBps ReadRequests=%dqps WriteRequests=%dqps\n",
			k, v.FsVfstype, v.BytesAll, v.BytesUsedPermillage, v.ReadBytes/uint64(sysinfo.PeriodSec)/1024, v.WriteBytes/uint64(sysinfo.PeriodSec)/1024,
			v.ReadRequests/uint64(sysinfo.PeriodSec), v.WriteRequests/uint64(sysinfo.PeriodSec))
	}

	fmt.Println("thermal:")
	t := sysinfo.Thermal
	for _, v := range t.Thermal {
		fmt.Printf("\t %s = %d\n", v.Type, v.Temp)
	}
	fmt.Printf("\n")
}

func main() {
	sm := s.New(1, callback)
	sm.OSEn = true
	sm.CPUEn = true
	sm.MemEn = true
	sm.NetEn = true
	sm.FsEn = true
	sm.ThermalEn = true

	sm.Start()
	defer sm.Stop()
	select {}
}
