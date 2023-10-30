package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/Senhnn/go_tool/shlarp"
	"golang.org/x/sys/unix"
	"math"
	"net"
	"net/netip"
	"time"
)

var (
	// ifaceFlag 设置发送ARP请求的网口。如：eth0
	ifaceFlag = flag.String("i", "eth0", "network interface to use for ARP request")

	// ipFlag 设置需要查询的IP地址
	ipFlag = flag.String("ip", "", "IPv4 address destination for ARP request")
)

func main() {
	flag.Parse()

	// 要查询的ip地址
	ip, err := netip.ParseAddr(*ipFlag)
	if err != nil {
		panic(err)
	}
	fmt.Println("Dst ip:", ip.String())

	// 查询指定的网口
	netIf, err := net.InterfaceByName(*ifaceFlag)
	if err != nil {
		panic(err)
	}

	// 使用arp包进行请求发送
	req, err := shlarp.NewIPv4ArpRequest(netIf, &ip)
	if err != nil {
		panic(err)
	}
	// AF_PACKET(packet socket)是一种Linux内核提供的用于直接访问网络数据包的接口。
	socket, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, unix.ETH_P_ARP)
	if err != nil {
		panic(err)
	}
	// 转换为网络字节序
	pnet, err := htons(unix.ETH_P_ARP)
	if err != nil {
		panic(err)
	}
	sa := &unix.SockaddrLinklayer{
		Protocol: pnet, // 以前是 直接写的 unix.ETH_P_ARP 忘记了转网络字节序
		Ifindex:  netIf.Index,
	}
	err = unix.Bind(socket, sa)
	if err != nil {
		panic(err)
	}
	marshal, err := req.Encode()
	fmt.Println(marshal)
	if err != nil {
		panic(err)
	}

	err = unix.Sendto(socket, marshal, 0, sa)
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second)

	recvBuf := make([]byte, 1000, 1000)
	n, _, _ := unix.Recvfrom(socket, recvBuf, 0)
	reply := req
	fmt.Println(recvBuf[:n])
	err = reply.Decode(recvBuf[:n])
	if err != nil {
		panic(err)
	}
	mac := net.HardwareAddr(reply.SourceHardwareAddress[:])
	fmt.Printf("%s -> %s\n", ip, mac)
}

func htons(i int) (uint16, error) {
	if i < 0 || i > math.MaxUint16 {
		return 0, errors.New("网络字节序错误")
	}

	// 大端方式保存
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(i))

	// 转换为网络字节序
	return binary.NativeEndian.Uint16(b[:]), nil
}
