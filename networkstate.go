package systemmonitor

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	netif_ignore = []string{
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
	ERROR_STATUS    string = "00"
	TCP_ESTABLISHED string = "01"
	TCP_SYN_SENT    string = "02"
	TCP_SYN_RECV    string = "03"
	TCP_FIN_WAIT1   string = "04"
	TCP_FIN_WAIT2   string = "05"
	TCP_TIME_WAIT   string = "06"
	TCP_CLOSE       string = "07"
	TCP_CLOSE_WAIT  string = "08"
	TCP_LAST_ACK    string = "09"
	TCP_LISTEN      string = "0A"
	TCP_CLOSING     string = "0B"
)

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

	pre_TotalInBytes     uint64
	pre_TotalInPackages  uint64
	pre_TotalOutBytes    uint64
	pre_TotalOutPackages uint64
}

type TcpInfo struct {
	Tcp_error_status uint64
	Tcp_established  uint64
	Tcp_syn_sent     uint64
	Tcp_syn_recv     uint64
	Tcp_fin_wait1    uint64
	Tcp_fin_wait2    uint64
	Tcp_time_wait    uint64
	Tcp_close        uint64
	Tcp_close_wait   uint64
	Tcp_last_ack     uint64
	Tcp_listen       uint64
	Tcp_closing      uint64
	TcpConnections   uint64 //total
}

type NetWorkInfo struct {
	Cards map[string]*NetIf
	Tcp   *TcpInfo
	Tcp6  *TcpInfo
}

func _gettcpcount(path string) (TcpInfo, error) {
	var tcp TcpInfo
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
		case ERROR_STATUS:
			tcp.Tcp_error_status += 1
		case TCP_ESTABLISHED:
			tcp.Tcp_established += 1
		case TCP_SYN_SENT:
			tcp.Tcp_syn_sent += 1
		case TCP_SYN_RECV:
			tcp.Tcp_syn_recv += 1
		case TCP_FIN_WAIT1:
			tcp.Tcp_fin_wait1 += 1
		case TCP_FIN_WAIT2:
			tcp.Tcp_fin_wait2 += 1
		case TCP_TIME_WAIT:
			tcp.Tcp_time_wait += 1
		case TCP_CLOSE:
			tcp.Tcp_close += 1
		case TCP_CLOSE_WAIT:
			tcp.Tcp_close_wait += 1
		case TCP_LAST_ACK:
			tcp.Tcp_last_ack += 1
		case TCP_LISTEN:
			tcp.Tcp_listen += 1
		case TCP_CLOSING:
			tcp.Tcp_closing += 1
		}
		tcp.TcpConnections += 1
	}
	file.Close()
	return tcp, nil
}

func gettcpcount() (TcpInfo, TcpInfo, error) {
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

func (this *NetWorkInfo) getnetif() error {
	if this.Cards == nil {
		this.Cards = make(map[string]*NetIf)
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
		for _, ignore := range netif_ignore {
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
		if _, ok := this.Cards[name]; !ok {
			this.Cards[name] = &NetIf{}
			this.Cards[name].Iface = name
		}
		netcard := this.Cards[name]

		netcard.TotalInBytes = inbytes
		netcard.TotalInPackages = inpackages
		netcard.TotalOutBytes = outbytes
		netcard.TotalOutPackages = outpackages

		if netcard.pre_TotalInBytes > 0 && netcard.pre_TotalInPackages > 0 {
			netcard.InBytes = inbytes - netcard.pre_TotalInBytes
			netcard.InPackages = inpackages - netcard.pre_TotalInPackages
		}
		if netcard.pre_TotalOutBytes > 0 && netcard.pre_TotalOutPackages > 0 {
			netcard.OutBytes = outbytes - netcard.pre_TotalOutBytes
			netcard.OutPackages = outpackages - netcard.pre_TotalOutPackages
		}

		netcard.pre_TotalInBytes = inbytes
		netcard.pre_TotalInPackages = inpackages
		netcard.pre_TotalOutBytes = outbytes
		netcard.pre_TotalOutPackages = outpackages
	}
	file.Close()
	return nil
}

func (this *NetWorkInfo) getnetworkstate() error {
	err := this.getnetif()
	if err != nil {
		return err
	}
	tcpcount, tcpcount6, errtcp := gettcpcount()
	if errtcp != nil {
		return errtcp
	}
	this.Tcp = &tcpcount
	this.Tcp6 = &tcpcount6

	return nil
}
