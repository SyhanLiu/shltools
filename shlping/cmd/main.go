package main

import (
	"flag"
	"fmt"
	"github.com/Senhnn/go_tool/shlping"
	"os"
	"os/signal"
	"time"
)

var usage = `
用法:
    ping [-c count] [-t timeout] host
样例:
	-l：设置TTL（默认64）
	-i：ping间隔时间（单位为ms）
    # 持续ping
    ping www.google.com

    # ping5次
    ping -c 5 www.google.com

    # ping并且设置10秒超时
    ping -t 10s www.google.com
`

func main() {
	timeout := flag.Duration("t", time.Second*10, "")
	count := flag.Int("c", -1, "")
	interval := flag.Int("i", 1000, "")
	ttl := flag.Int("l", 64, "TTL")

	flag.Usage = func() {
		fmt.Print(usage)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}

	host := flag.Arg(0)
	pinger, err := shlping.NewPinger(host)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	// listen for ctrl-C signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pinger.Stop()
		}
	}()

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}
	pinger.OnDuplicateRecv = func(pkt *ping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %d duplicates, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketsRecvDuplicates, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	pinger.Count = *count
	pinger.Size = *size
	pinger.Interval = *interval
	pinger.Timeout = *timeout
	pinger.TTL = *ttl
	pinger.SetPrivileged(*privileged)

	fmt.Printf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	err = pinger.Run()
	if err != nil {
		fmt.Println("Failed to ping target host:", err)
	}
}
