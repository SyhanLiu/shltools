package shlnl

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
)

func WriteNlMsghdrToBuf(p *unix.NlMsghdr) []byte {
	var buf []byte = make([]byte, unix.SizeofNlMsghdr)
	var i, l = 0, 0
	l = binary.Size(p.Len)
	binary.BigEndian.PutUint32(buf[i:i+l], p.Len)
	i += l
	l = binary.Size(p.Type)
	binary.BigEndian.PutUint16(buf[i:i+l], p.Type)
	i += l
	l = binary.Size(p.Flags)
	binary.BigEndian.PutUint16(buf[i:i+l], p.Flags)
	i += l
	l = binary.Size(p.Seq)
	binary.BigEndian.PutUint32(buf[i:i+l], p.Seq)
	i += l
	l = binary.Size(p.Pid)
	binary.BigEndian.PutUint32(buf[i:i+l], p.Pid)
	return buf
}

func WriteIfInfomsgToBuf(p *unix.IfInfomsg) []byte {
	var buf []byte = make([]byte, unix.SizeofIfInfomsg)
	var i, l = 0, 0
	buf[i] = p.Family
	i += 1
	i += 1 // 跳过pad
	l = binary.Size(p.Type)
	binary.BigEndian.PutUint16(buf[i:i+l], p.Type)
	i += l
	l = binary.Size(p.Index)
	binary.BigEndian.PutUint32(buf[i:i+l], uint32(p.Index))
	i += l
	l = binary.Size(p.Flags)
	binary.BigEndian.PutUint32(buf[i:i+l], p.Flags)
	i += l
	l = binary.Size(p.Change)
	binary.BigEndian.PutUint32(buf[i:i+l], p.Change)
	return buf
}

func WriteRtAttrToBuf(p *unix.RtAttr, b []byte) []byte {
	var buf []byte = make([]byte, unix.SizeofRtAttr)
	var i, l = 0, 0
	l = binary.Size(p.Len)
	binary.BigEndian.PutUint16(buf[i:i+l], p.Len)
	i += l
	l = binary.Size(p.Type)
	binary.BigEndian.PutUint16(buf[i:i+l], p.Type)
	buf = append(buf, b...)
	return buf
}
