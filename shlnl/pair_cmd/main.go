package main

import (
	"fmt"
	"github.com/Senhnn/go_tool/shlnl"
	"golang.org/x/sys/unix"
	"time"
	"unsafe"
)

var veth = "veth1"
var pveth = "pveth1"

func main() {
	// 常见netlink套接字
	socket, err := shlnl.NlSocket(unix.NETLINK_ROUTE)
	if err != nil {
		panic(fmt.Sprintf("Create net link sock error:%s", err.Error()))
	}

	nlSockAddr := &unix.SockaddrNetlink{Family: unix.AF_NETLINK}
	data := make([]byte, 0)

	nlMsgHdr := &unix.NlMsghdr{
		Len: unix.SizeofNlMsghdr + unix.SizeofIfInfomsg, // 消息头的长度
		// RTM_NEWLINK, RTM_DELLINK, RTM_GETLINK。创建，删除，获取特定的网络设备
		Type: unix.RTM_NEWLINK,
		// NLM_F_REQUEST：请求消息，从用户空间到内核空间的消息都需要将该位置位。
		// NLM_F_CREATE：Create object if it doesn't already exist. -- From linux man
		// NLM_F_EXCL：Don't replace if the object already exists.  -- From linux man
		// NLM_F_ACK：要求内核为该请求发送回复响应。Request for an acknowledgement on success. -- From linux man
		Flags: unix.NLM_F_REQUEST | unix.NLM_F_CREATE | unix.NLM_F_EXCL | unix.NLM_F_ACK,
		Seq:   uint32(time.Now().Unix()),
		Pid:   0,
	}

	ifInfoMsg := &unix.IfInfomsg{
		// AF_UNSPEC:函数返回的是适用于指定主机名和服务名且适合任何协议族的地址。
		Family: unix.AF_UNSPEC,
	}
	data = append(data, shlnl.WriteIfInfomsgToBuf(ifInfoMsg)...)

	rta := &unix.RtAttr{
		Len:  unix.SizeofRtAttr + uint16(len(veth)+1),
		Type: unix.IFLA_IFNAME, // 表示指定名称
	}
	data = append(data, shlnl.WriteRtAttrToBuf(rta, append([]byte(veth), 0))...)

	rtaLinkInfo := &unix.RtAttr{
		Len:  unix.SizeofRtAttr,
		Type: unix.IFLA_LINKINFO, // 设置link info
	}
	data = append(data, shlnl.WriteRtAttrToBuf(rtaLinkInfo, nil)...)

	rtaInfoKind := &unix.RtAttr{
		Len:  unix.SizeofRtAttr + uint16(len("veth")) + 1,
		Type: unix.IFLA_INFO_KIND, // 指定类型，此处最终会根据"veth"找到操作集
	}
	/*
		// 详情查看linux内核源码
		file:drivers\net\veth.c
		#define DRV_NAME	"veth"
		static struct rtnl_link_ops veth_link_ops = {
			.kind		= DRV_NAME,
			.priv_size	= sizeof(struct veth_priv),
			.setup		= veth_setup,
			.validate	= veth_validate,
			.newlink	= veth_newlink,
			.dellink	= veth_dellink,
			.policy		= veth_policy,
			.maxtype	= VETH_INFO_MAX,
			.get_link_net	= veth_get_link_net,
			.get_num_tx_queues	= veth_get_num_queues,
			.get_num_rx_queues	= veth_get_num_queues,
		};
	*/
	data = append(data, shlnl.WriteRtAttrToBuf(rtaInfoKind, append([]byte("veth"), 0))...)

	rtaInfoData := &unix.RtAttr{
		Len:  unix.SizeofRtAttr,
		Type: unix.IFLA_INFO_DATA,
	}
	data = append(data, shlnl.WriteRtAttrToBuf(rtaInfoData, nil)...)

	rtaInfoPeer := &unix.RtAttr{
		Len:  unix.SizeofRtAttr,
		Type: 1, // 对端接口
	}
	data = append(data, shlnl.WriteRtAttrToBuf(rtaInfoPeer, nil)...)

	rta = &unix.RtAttr{
		Len:  unix.SizeofRtAttr + uint16(len(pveth)+1),
		Type: unix.IFLA_IFNAME, // 表示指定名称
	}
	data = append(data, shlnl.WriteRtAttrToBuf(rta, append([]byte(pveth), 0))...)

	nlMsgHdr.Len += uint32(rta.Len + rtaLinkInfo.Len + rtaInfoKind.Len + rtaInfoData.Len + rtaInfoPeer.Len + rta.Len)
	fmt.Println(nlMsgHdr.Len)
	data = append(shlnl.WriteNlMsghdrToBuf(nlMsgHdr), data...)

	err = unix.Sendto(socket, data, 0, nlSockAddr)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 4096)
	_, _, err = unix.Recvfrom(socket, buf, 0)
	if err != nil {
		panic(err)
	}
	unix.Close(socket)

	uptr := uintptr(unsafe.Pointer(&buf[0]))
	ret := (*unix.NlMsghdr)(unsafe.Pointer(&buf[0]))
	if ret.Type == unix.NLMSG_ERROR {
		nlerr := (*unix.NlMsgerr)(unsafe.Pointer(uptr + unix.SizeofNlMsghdr))
		if nlerr.Error < 0 {
			panic(fmt.Sprintf("error: %d\n, failed to create links\n", nlerr.Error))
		}
	} else {
		panic("failed to create links")
	}
}
