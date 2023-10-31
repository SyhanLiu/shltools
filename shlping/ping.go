package shlping

import (
	"github.com/google/uuid"
	"net"
	"sync"
	"time"
)

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

	// 往返时间统计
	minRtt    time.Duration
	maxRtt    time.Duration
	avgRtt    time.Duration
	stdDevRtt time.Duration
	stddevm2  time.Duration
	statsMu   sync.RWMutex

	// rtts 所有包的RTT
	rtts []time.Duration

	// OnSetup Pinger has finished setting up the listening socket
	OnSetup func()
	// OnSend Pinger发送数据时触发
	OnSend func(*Packet)
	// OnRecv Pinger接收数据时触发
	OnRecv func(*Packet)
	// OnFinish Pinger完成时触发
	OnFinish func(*Statistics)
	// OnDuplicateRecv Pinger重复收到数据包时触发
	OnDuplicateRecv func(*Packet)

	// TTL 跳数
	TTL int
	// Size 数据包的大小
	Size int
	// Tracker: Used to uniquely identify packets - Deprecated
	Tracker uint64
	// Source 源IP
	Source string
	// 完成标志
	done chan interface{}
	lock sync.Mutex
	// 目的地址
	ipaddr *net.IPAddr
	addr   string

	// trackerUUIDs 发包的uuid
	trackerUUIDs []uuid.UUID

	ipv4     bool
	id       int
	sequence int
	// awaitingSequences are in-flight sequence numbers we keep track of to help remove duplicate receipts
	awaitingSequences map[uuid.UUID]map[int]struct{}
	// network 为"ip","ip4","ip6"
	network string
	// protocol 为"icmp","udp"
	protocol string
}
