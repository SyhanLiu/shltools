package shlping

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
	"math"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	timeSliceLength  = 8
	trackerLength    = len(uuid.UUID{})
	protocolICMP     = unix.IPPROTO_ICMP
	protocolIPv6ICMP = unix.IPPROTO_ICMPV6
)

var (
	ipv4Proto = map[string]string{"icmp": "ip4:icmp", "udp": "udp4"}
	ipv6Proto = map[string]string{"icmp": "ip6:ipv6-icmp", "udp": "udp6"}
)

// NewPinger 新建一个Pinger
func NewPinger(addr string) (*Pinger, error) {
	p := newPinger(addr)
	return p, p.Resolve()
}

func newPinger(addr string) *Pinger {
	r := rand.New(rand.NewSource(getSeed()))
	firstUUID := uuid.New()
	var firstSequence = map[uuid.UUID]map[int]struct{}{}
	firstSequence[firstUUID] = make(map[int]struct{})
	return &Pinger{
		Interval:          time.Second,
		Timeout:           time.Duration(math.MaxInt64),
		Count:             -1,
		TTL:               64,
		Size:              0,
		lock:              sync.Mutex{},
		TargetAddr:        addr,
		trackerUUIDs:      []uuid.UUID{firstUUID},
		id:                r.Intn(math.MaxUint16),
		sequence:          0,
		awaitingSequences: firstSequence,
		network:           "ip",
		protocol:          "icmp",
	}
}

// Pinger 一个ping对象
type Pinger struct {
	// 两次发包时间间隔
	Interval time.Duration
	// 请求超时时间，超过该时间后回重新发包或者退出
	Timeout time.Duration
	// 发包次数
	Count int
	// 已经发送的包数
	PacketsSent int
	// 收到的包数
	PacketsRecv int
	// 收到重复的包数
	PacketsRecvDuplicates int

	// rtts 所有包的RTT
	rtts []time.Duration

	OnSetup func()
	// OnSend Pinger发送数据时触发
	OnSend func(*Packet)
	// OnRecv Pinger接收数据时触发
	OnRecv func(*Packet)
	// OnDuplicateRecv Pinger重复收到数据包时触发
	OnDuplicateRecv func(*Packet)

	// TTL 跳数
	TTL int
	// Size 数据包的大小
	Size int
	// Tracker 用于唯一表示数据包已经弃用
	Tracker uint64
	// 源地址
	SourceIpAddr *net.IPAddr
	SourceAddr   string
	lock         sync.Mutex
	// 目的地址
	TargetIpaddr *net.IPAddr
	TargetAddr   string

	// trackerUUIDs 发包的uuid
	trackerUUIDs []uuid.UUID

	id       int
	sequence int
	// awaitingSequences 记录sequence防止重复接受
	awaitingSequences map[uuid.UUID]map[int]struct{}
	// network 为"ip","ip4","ip6"
	network string
	// protocol 为"icmp","udp"
	protocol string
}

// Resolve 解析目的地址，如果是域名会由做域名解析
func (p *Pinger) Resolve() error {
	if len(p.TargetAddr) == 0 {
		return errors.New("addr cannot be empty")
	}
	ipaddr, err := net.ResolveIPAddr(p.network, p.TargetAddr)
	if err != nil {
		return err
	}
	// 是否是ipv4
	if len(ipaddr.IP.To4()) != net.IPv4len {
		panic("only support IPv4")
	}
	p.TargetIpaddr = ipaddr
	return nil
}

// Run 开始ping操作，会阻塞，可以使用Stop方法停止
func (p *Pinger) Run() error {
	var err error
	if p.TargetIpaddr == nil {
		err = p.Resolve()
	}
	if err != nil {
		return err
	}
	// 建立原始套接字
	sock := 0
	sock, err = unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_ICMP)
	if err != nil {
		fmt.Println(fmt.Sprintf("Create socket error:%s", err.Error()))
		return err
	}
	// 设置为手动写入ip首部
	err = unix.SetsockoptInt(sock, unix.IPPROTO_IP, unix.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println(fmt.Sprintf("Set socket IP_HDRINCL error:%s", err.Error()))
		return err
	}
	// 解析本地源IP地址
	p.SourceAddr = "192.168.0.101"
	p.SourceIpAddr, err = net.ResolveIPAddr(p.network, p.SourceAddr)
	if err != nil {
		fmt.Println(fmt.Sprintf("ResolveIPAddr SourceAddr:%s error:%s", p.SourceAddr, err.Error()))
		return err
	}
	sa := &unix.SockaddrInet4{}
	copy(sa.Addr[:], p.SourceIpAddr.IP)
	// 绑定本地源IP地址
	err = unix.Bind(sock, sa)
	if err != nil {
		fmt.Println(fmt.Sprintf("Bind SourceAddr:%s error:%s", p.SourceIpAddr.String(), err.Error()))
		return err
	}

	if handler := p.OnSetup; handler != nil {
		handler()
	}
	err = p.sendICMP(sock)
	if err != nil {
		return err
	}
	for {
		data, err := p.recvICMP(sock)
		if err != nil {
			return err
		}
		if data.IPv4Header.Src.String() == p.TargetIpaddr.String() {
			if data.ICMPData.Type == ipv4.ICMPTypeEchoReply {
				fmt.Println(fmt.Sprintf("ip:%s reply", data.IPv4Header.Src.String()))
				break
			}
		} else {
			fmt.Println(fmt.Sprintf("other data ip:%s type:%d", data.IPv4Header.Src.String(), data.IPv4Header.Protocol))
		}
	}
	return err
}

func (p *Pinger) sendICMP(sock int) error {
	icmpData := &icmp.Message{
		Type:     ipv4.ICMPTypeEcho,
		Code:     0,
		Checksum: 0,
		Body: &icmp.Echo{
			ID:   0,
			Seq:  0,
			Data: nil,
		},
	}
	buff, err := icmpData.Marshal(nil)
	if err != nil {
		return err
	}
	data := &ICMPv4Data{
		IPv4Header: &ipv4.Header{
			Version: ipv4.Version,
			Len:     ipv4.HeaderLen, // IP头长一般是20
			TOS:     0x00,
			//buff为数据
			TotalLen: ipv4.HeaderLen + len(buff),
			TTL:      64,
			Flags:    ipv4.DontFragment, // 不分片
			FragOff:  0,
			Protocol: unix.IPPROTO_ICMP,
			Checksum: 0,
			Src:      p.SourceIpAddr.IP,
			Dst:      p.TargetIpaddr.IP,
		},
		ICMPData: icmpData,
	}
	marshal, err := data.Marshal()
	if err != nil {
		return err
	}
	// 解析目的IP地址
	peer, err := net.ResolveIPAddr(p.network, p.TargetAddr)
	if err != nil {
		fmt.Println(fmt.Sprintf("ResolveIPAddr TargetAddr:%s error:%s", p.TargetAddr, err.Error()))
		return err
	}
	sa := &unix.SockaddrInet4{}
	copy(sa.Addr[:], peer.IP)
	err = unix.Sendto(sock, marshal, 0, sa)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pinger) recvICMP(sock int) (*ICMPv4Data, error) {
	for {
		select {
		default:
			bytes := make([]byte, 4096)
			n, _, err := unix.Recvfrom(sock, bytes, 0)
			if err != nil {
				continue
			}
			data := &ICMPv4Data{
				IPv4Header: &ipv4.Header{},
				ICMPData:   nil,
			}
			// 解析icmp报文
			err = data.Unmarshal(bytes[:n])
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}
}

var seed int64 = time.Now().UnixNano()

// getSeed returns a goroutine-safe unique seed
func getSeed() int64 {
	return atomic.AddInt64(&seed, 1)
}
