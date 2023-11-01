package main

import (
	"flag"
	"fmt"
	"github.com/Senhnn/go_tool/shlping"
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
	//timeout := flag.Duration("t", time.Second*10, "")
	//count := flag.Int("c", -1, "")
	//interval := flag.Int("i", 1000, "")
	//ttl := flag.Int("l", 64, "TTL")

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

	pinger.OnRecv = func(pkt *shlping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}
	pinger.OnDuplicateRecv = func(pkt *shlping.Packet) {
		fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v ttl=%v (DUP!)\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt, pkt.Ttl)
	}

	//pinger.Count = *count
	//pinger.Interval = time.Duration(*interval)
	//pinger.Timeout = (*timeout) * time.Millisecond
	//pinger.TTL = *ttl

	fmt.Printf("PING %s (%s):\n", pinger.TargetAddr, pinger.TargetIpaddr)
	err = pinger.Run()
	if err != nil {
		fmt.Println("Failed to ping target host:", err)
	}
}
