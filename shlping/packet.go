package shlping

import (
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
	"net"
	"time"
)

// 数据包
type packet struct {
	bytes  []byte
	nbytes int
	ttl    int
}

// Packet 用于表示ICMP包
type Packet struct {
	// Rtt ping的往返时间
	Rtt time.Duration
	// IPAddr 目的地址
	IPAddr *net.IPAddr
	// Addr 目的地址
	Addr string
	// NBytes 数据长度
	Nbytes int
	// Seq icmp包的序号
	Seq int
	// TTL 包的TTL
	Ttl int
	// ID ICMP包的唯一ID
	ID int
}

type ICMPv4Data struct {
	IPv4Header *ipv4.Header
	ICMPData   *icmp.Message
}

func (i *ICMPv4Data) Marshal() ([]byte, error) {
	var err error
	res := make([]byte, 4096)
	ipHeader, err := i.IPv4Header.Marshal()
	if err != nil {
		fmt.Println(fmt.Sprintf("ip header marshal error:%s:", err.Error()))
		return nil, err
	}
	copy(res[:], ipHeader)

	icmpData, err := i.ICMPData.Marshal(nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("icmpData marshal error:%s:", err.Error()))
		return nil, err
	}
	copy(res[len(ipHeader):], icmpData)
	return res, nil
}

func (i *ICMPv4Data) Unmarshal(b []byte) error {
	var err error
	err = i.IPv4Header.Parse(b[:ipv4.HeaderLen])
	if err != nil {
		fmt.Println(fmt.Sprintf("ipv4 header unmarshal error:%s:", err.Error()))
		return err
	}

	icmpData, err := icmp.ParseMessage(unix.IPPROTO_ICMP, b[ipv4.HeaderLen:])
	if err != nil {
		fmt.Println(fmt.Sprintf("icmpData unmarshal error:%s:", err.Error()))
		return err
	}
	i.ICMPData = icmpData
	return nil
}
