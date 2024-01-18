package shlnl

import "golang.org/x/sys/unix"

// NlSocket 封装netlink套接字
func NlSocket(proto int) (int, error) {
	return unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, proto)
}
