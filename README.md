# systemmonitor Quickstart Guide
---

systemmonitor是go语言编写的“系统资源监视器”，监视的资源包括：系统信息、CPU使用、内存使用、网络、硬盘IO、主板温度

用法可以参考[这个简单的web项目](https://github.com/lambda-zhang/systemmonitor-web)

注意： 目前只支持linux


# how to run example
---
```
$ go get -u -v github.com/lambda-zhang/systemmonitor
$ cd $GOPATH/src/github.com/lambda-zhang/systemmonitor
$ go run example/test_monitor.go
###############################################
os:
	StartTime=2019-02-16 08:46:40 +0800 CST Arch=amd64 Os=linux KernelVersion=Linux-4.4.0-131-generic KernelHostname=lambda-Lenovo NumCpu=8 UpTime=36701 UsePermillage=35‰
cpu:
	Cpu_permillage=27‰  Avg1min=0.840000
memory:
	MemTotal=16779108352  MemUsePermillage=228‰  SwapTotal=2046816256  SwapUsePermillage=0‰
net:
	eth1: inKBps=0 outKBps=0
	eth0: inKBps=0 outKBps=0
	total TcpConnections=91 ESTABLISHED=42 TCP_LISTEN=34
disk:
	sda7 FsVfstype=TODO BytesAll=257588736000 BytesUsedPermillage=822‰ ReadBytes=0KBps  WriteBytes=124KBps ReadRequests=0qps WriteRequests=2qps
	sda2 FsVfstype=ext4 BytesAll=105192407040 BytesUsedPermillage=825‰ ReadBytes=0KBps  WriteBytes=0KBps ReadRequests=0qps WriteRequests=0qps
	sda5 FsVfstype=ext4 BytesAll=126692069376 BytesUsedPermillage=932‰ ReadBytes=0KBps  WriteBytes=0KBps ReadRequests=0qps WriteRequests=0qps
thermal:
	 acpitz0 = 27800
	 acpitz1 = 29800
	 x86_pkg_temp2 = 34000
```
