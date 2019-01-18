package main

import (
	"fmt"
	"time"

	s "github.com/lambda-zhang/systemmonitor"
)

var period_sec int = 1

func callback(sysinfo *s.SysInfo) {
	fmt.Println("###############################################")
	fmt.Println("os:")
	o := sysinfo.OS
	fmt.Printf("\tStartTime=%s Arch=%s Os=%s KernelVersion=%s KernelHostname=%s NumCpu=%d UpTime=%d UsePermillage=%d‰\n",
		time.Unix(o.StartTime, 0), o.Arch, o.Os, o.KernelVersion, o.KernelHostname, o.NumCpu, o.UpTime, o.UsePermillage)

	fmt.Println("cpu:")
	c := sysinfo.CPU
	fmt.Printf("\tCpu_permillage=%d‰  Avg1min=%f\n", c.Cpu_permillage, c.Avg1min)

	fmt.Println("memory:")
	m := sysinfo.Mem
	fmt.Printf("\tMemTotal=%d  MemUsePermillage=%d‰  SwapTotal=%d  SwapUsePermillage=%d‰\n", m.MemTotal, m.MemUsePermillage, m.SwapTotal, m.SwapUsePermillage)

	fmt.Println("net:")
	n := sysinfo.Net
	for k, v := range n.Cards {
		fmt.Printf("\t%s: inKBps=%d outKBps=%d\n", k, v.InBytes/uint64(period_sec)/1024, v.OutBytes/uint64(period_sec)/1024)
	}
	fmt.Printf("\ttotal TcpConnections=%d ESTABLISHED=%d TCP_LISTEN=%d\n",
		n.Tcp.TcpConnections+n.Tcp6.TcpConnections, n.Tcp.Tcp_established+n.Tcp6.Tcp_established, n.Tcp.Tcp_listen+n.Tcp6.Tcp_listen)

	fmt.Println("disk:")
	f := sysinfo.Fs
	for k, v := range f.Disks {
		fmt.Printf("\t%s FsVfstype=%s BytesAll=%d BytesUsedPermillage=%d‰ ReadBytes=%dKBps  WriteBytes=%dKBps ReadRequests=%dqps WriteRequests=%dqps\n",
			k, v.FsVfstype, v.BytesAll, v.BytesUsedPermillage, v.ReadBytes/uint64(period_sec)/1024, v.WriteBytes/uint64(period_sec)/1024, v.ReadRequests/uint64(period_sec), v.WriteRequests/uint64(period_sec))
	}

	fmt.Println("thermal:")
	t := sysinfo.Thermal
	t.GetThermal()
	for _, v := range t.Thermal {
		fmt.Printf("\t %s = %d\n", v.Type, v.Temp)
	}
	fmt.Println("\n\n\n\n")
}

func main() {
	sm := s.New(period_sec, callback)
	sm.Start()
	defer sm.Stop()
	select {}
}
