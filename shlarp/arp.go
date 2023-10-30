package shlarp

import (
	"encoding/binary"
	"errors"
	"net"
	"net/netip"
)

// arp请求为1；arp应答为2；rarp请求为3；rarp应答为4。
const (
	ARPRequest  = 1
	ARPReply    = 2
	RARPRequest = 3
	RARPReply   = 4
)

// protocolARP 对应MAC头的类型字段
const protocolARP = 0x0806

// EthernetHeader MAC（以太网）头部
type EthernetHeader struct {
	Dst     [6]byte // 目的地的mac地址
	Src     [6]byte // 本地以太网接口的mac地址
	EthType uint16  // 长度或者类型
}

// ArpIPv4Header ARP arp的头部
type ArpIPv4Header struct {
	EthernetHeader
	HardwareType          uint16 // 对于以太网该值为1
	ProtocolType          uint16 // 对于IPv4地址，该值为0x0800
	HardwareSize          uint8  // 以太网中使用IPv4地址的ARP请求或应答，该值为6
	ProtocolSize          uint8  // 以太网中使用IPv4地址的ARP请求或应答，该值为4
	Op                    uint16 // 指出该操作，arp请求为1；arp应答为2；rarp请求为3；rarp应答为4
	SourceHardwareAddress [6]byte
	SourceProtocolAddress [4]byte
	DstHardwareAddress    [6]byte
	DstProtocolAddress    [4]byte
	Padding               [18]byte
}

// Encode 序列化
func (header *ArpIPv4Header) Encode() ([]byte, error) {
	res := make([]byte, binary.Size(header), binary.Size(header))
	addr := 0
	copy(res[addr:], header.Dst[:])
	addr += len(header.Dst)
	copy(res[addr:], header.Src[:])
	addr += len(header.Src)
	binary.BigEndian.PutUint16(res[addr:], header.EthType)
	addr += binary.Size(header.EthType)
	binary.BigEndian.PutUint16(res[addr:], header.HardwareType)
	addr += binary.Size(header.HardwareType)
	binary.BigEndian.PutUint16(res[addr:], header.ProtocolType)
	addr += binary.Size(header.ProtocolType)
	copy(res[addr:], string(header.HardwareSize))
	addr += binary.Size(header.HardwareSize)
	copy(res[addr:], string(header.ProtocolSize))
	addr += binary.Size(header.ProtocolSize)
	binary.BigEndian.PutUint16(res[addr:], header.Op)
	addr += binary.Size(header.Op)
	copy(res[addr:], header.SourceHardwareAddress[:])
	addr += len(header.SourceHardwareAddress)
	copy(res[addr:], header.SourceProtocolAddress[:])
	addr += len(header.SourceProtocolAddress)
	copy(res[addr:], header.DstHardwareAddress[:])
	addr += len(header.DstHardwareAddress)
	copy(res[addr:], header.DstProtocolAddress[:])
	addr += len(header.DstProtocolAddress)
	copy(res[addr:], header.Padding[:])
	addr += len(header.Padding)

	return res, nil
}

// Decode 反序列化
func (header *ArpIPv4Header) Decode(raw []byte) error {
	addr := 0
	edge := addr + len(header.Dst)
	header.Dst = [6]byte(raw[addr:edge])
	addr = edge
	edge = addr + len(header.Dst)
	header.Src = [6]byte(raw[addr:edge])
	addr = edge
	edge = addr + binary.Size(header.EthType)
	header.EthType = binary.BigEndian.Uint16(raw[addr:edge])
	addr = edge
	edge = addr + binary.Size(header.HardwareType)
	header.HardwareType = binary.BigEndian.Uint16(raw[addr:edge])
	addr = edge
	edge = addr + binary.Size(header.ProtocolType)
	header.ProtocolType = binary.BigEndian.Uint16(raw[addr:edge])
	addr = edge
	edge = addr + binary.Size(header.HardwareSize)
	header.HardwareSize = raw[addr]
	addr = edge
	edge = addr + binary.Size(header.ProtocolSize)
	header.ProtocolSize = raw[addr]
	addr = edge
	edge = addr + binary.Size(header.Op)
	header.Op = binary.BigEndian.Uint16(raw[addr:edge])
	addr = edge
	edge = addr + len(header.SourceHardwareAddress)
	header.SourceHardwareAddress = [6]byte(raw[addr:edge])
	addr = edge
	edge = addr + len(header.SourceProtocolAddress)
	header.SourceProtocolAddress = [4]byte(raw[addr:edge])
	addr = edge
	edge = addr + len(header.DstHardwareAddress)
	header.DstHardwareAddress = [6]byte(raw[addr:edge])
	addr = edge
	edge = addr + len(header.DstProtocolAddress)
	header.DstProtocolAddress = [4]byte(raw[addr:edge])
	addr = edge
	edge = addr + len(header.Padding)
	header.Padding = [18]byte(raw[addr:edge])

	return nil
}

// NewIPv4ArpRequest 新建一个IPv4协议下的arp头部
func NewIPv4ArpRequest(netIf *net.Interface, dstIp *netip.Addr) (*ArpIPv4Header, error) {
	header := &ArpIPv4Header{}
	localAddrs, _ := netIf.Addrs()
	if len(localAddrs) == 0 {
		return nil, errNoIPv4Addr
	}
	addr, _, err := net.ParseCIDR(localAddrs[0].String())
	if err != nil {
		return nil, err
	}
	localAddr, err := netip.ParseAddr(addr.String())
	if err != nil {
		return nil, err
	}
	localAddr.As4()

	localMac := netIf.HardwareAddr

	// MAC头
	header.Dst = [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff} // 广播地址
	header.Src = [6]byte(localMac)
	header.EthType = protocolARP // 固定数值，表示arp
	// ARP头
	header.HardwareType = 1     // 以太网
	header.ProtocolType = 0x800 // ipv4
	header.HardwareSize = 6
	header.ProtocolSize = 4
	header.Op = ARPRequest // arp请求为1
	header.SourceHardwareAddress = [6]byte(localMac)
	header.SourceProtocolAddress = localAddr.As4()
	header.DstHardwareAddress = [6]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	header.DstProtocolAddress = dstIp.As4()
	header.Padding = [18]byte{}
	return header, nil
}

// errNoIPv4Addr 网口没有ipv4地址
var errNoIPv4Addr = errors.New("no IPv4 address available for interface")
