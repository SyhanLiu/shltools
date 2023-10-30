package main

import (
	"encoding/binary"
	"errors"
	"math"
)

func main() {
	u, _ := htons(0x806)
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(u))
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
