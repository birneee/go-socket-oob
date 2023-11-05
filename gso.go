package go_udp

import (
	"golang.org/x/sys/unix"
	"syscall"
	"unsafe"
)

func appendUDPSegmentSizeMsg(oob []byte, size uint16) []byte {
	const dataLen = 2 // payload is a uint16
	oob, cmsgData := appendCmsg(oob, syscall.IPPROTO_UDP, unix.UDP_SEGMENT, dataLen)
	*(*uint16)(unsafe.Pointer(&cmsgData[0])) = size
	return oob
}
