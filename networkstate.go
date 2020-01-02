package systemmonitor

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	netifIgnore = []string{
		"sit",
		"dummy",
		"docker",
		"br-",
		"lo",
		"lxcbr",
		"veth",
	}
)

const (
	errorStatus    string = "00"
	tcpEstablished string = "01"
	tcpSynSent     string = "02"
	tcpSynRecv     string = "03"
	tcpFinWait1    string = "04"
	tcpFinWait2    string = "05"
	tcpTimewait    string = "06"
	tcpClose       string = "07"
	tcpCloseWait   string = "08"
	tcpLastAck     string = "09"
	tcpListen      string = "0A"
	tcpClosing     string = "0B"
)

// NetIf 网卡信息
type NetIf struct {
	Iface string

	InBytes         uint64
	InPackages      uint64
	TotalInBytes    uint64
	TotalInPackages uint64

	OutBytes         uint64
	OutPackages      uint64
	TotalOutBytes    uint64
	TotalOutPackages uint64

	preTotalInBytes     uint64
	preTotalInPackages  uint64
	preTotalOutBytes    uint64
	preTotalOutPackages uint64
}

// TCPInfo tcp信息
type TCPInfo struct {
	TCPErrorStatus uint64
	TCPEstablished uint64
	TCPSynSent     uint64
	TCPSynRecv     uint64
	TCPFinWait1    uint64
	TCPFinWait2    uint64
	TCPTimewait    uint64
	TCPClose       uint64
	TCPCloseWait   uint64
	TCPLastAck     uint64
	TCPListen      uint64
	TCPClosing     uint64
	TCPConnections uint64 //total
}

// NetWorkInfo 网卡信息
type NetWorkInfo struct {
	Cards map[string]*NetIf
	TCP   *TCPInfo
	TCP6  *TCPInfo
}

func _gettcpcount(path string) (TCPInfo, error) {
	var tcp TCPInfo
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) != 17 || fields[3] == "" {
			continue
		}
		switch fields[3] {
		case errorStatus:
			tcp.TCPErrorStatus++
		case tcpEstablished:
			tcp.TCPEstablished++
		case tcpSynSent:
			tcp.TCPSynSent++
		case tcpSynRecv:
			tcp.TCPSynRecv++
		case tcpFinWait1:
			tcp.TCPFinWait1++
		case tcpFinWait2:
			tcp.TCPFinWait2++
		case tcpTimewait:
			tcp.TCPTimewait++
		case tcpClose:
			tcp.TCPClose++
		case tcpCloseWait:
			tcp.TCPCloseWait++
		case tcpLastAck:
			tcp.TCPLastAck++
		case tcpListen:
			tcp.TCPListen++
		case tcpClosing:
			tcp.TCPClosing++
		}
		tcp.TCPConnections++
	}
	file.Close()
	return tcp, nil
}

func gettcpcount() (TCPInfo, TCPInfo, error) {
	tcpcount, err := _gettcpcount("/proc/net/tcp")
	if err != nil {
		return tcpcount, tcpcount, err
	}
	tcpcount6, err6 := _gettcpcount("/proc/net/tcp6")
	if err6 != nil {
		return tcpcount, tcpcount6, err6
	}

	return tcpcount, tcpcount6, nil
}

func (ninfo *NetWorkInfo) getnetif() error {
	if ninfo.Cards == nil {
		ninfo.Cards = make(map[string]*NetIf)
	}
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		panic(err)
	}
	rd := bufio.NewReader(file)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) != 17 || fields[0] == "" {
			continue
		}
		isignored := false
		for _, ignore := range netifIgnore {
			if strings.Index(fields[0], ignore) == 0 {
				isignored = true
			}
		}
		if isignored == true {
			continue
		}

		name := strings.Trim(fields[0], ":")
		if name == "" {
			continue
		}
		inbytes, _ := strconv.ParseUint(fields[1], 10, 64)
		inpackages, _ := strconv.ParseUint(fields[2], 10, 64)
		outbytes, _ := strconv.ParseUint(fields[9], 10, 64)
		outpackages, _ := strconv.ParseUint(fields[10], 10, 64)
		if _, ok := ninfo.Cards[name]; !ok {
			ninfo.Cards[name] = &NetIf{}
			ninfo.Cards[name].Iface = name
		}
		netcard := ninfo.Cards[name]

		netcard.TotalInBytes = inbytes
		netcard.TotalInPackages = inpackages
		netcard.TotalOutBytes = outbytes
		netcard.TotalOutPackages = outpackages

		if netcard.preTotalInBytes > 0 && netcard.preTotalInPackages > 0 {
			netcard.InBytes = inbytes - netcard.preTotalInBytes
			netcard.InPackages = inpackages - netcard.preTotalInPackages
		}
		if netcard.preTotalOutBytes > 0 && netcard.preTotalOutPackages > 0 {
			netcard.OutBytes = outbytes - netcard.preTotalOutBytes
			netcard.OutPackages = outpackages - netcard.preTotalOutPackages
		}

		netcard.preTotalInBytes = inbytes
		netcard.preTotalInPackages = inpackages
		netcard.preTotalOutBytes = outbytes
		netcard.preTotalOutPackages = outpackages
	}
	file.Close()
	return nil
}

func (ninfo *NetWorkInfo) getnetworkstate() error {
	err := ninfo.getnetif()
	if err != nil {
		return err
	}
	tcpcount, tcpcount6, errtcp := gettcpcount()
	if errtcp != nil {
		return errtcp
	}
	ninfo.TCP = &tcpcount
	ninfo.TCP6 = &tcpcount6

	return nil
}
