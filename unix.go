package go_udp

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

func appendCmsg(oob []byte, cmsgLevel int32, cmsgType int32, dataLen int) (newOob []byte, data []byte) {
	startLen := len(oob)
	oob = append(oob, make([]byte, unix.CmsgLen(dataLen))...)
	h := (*unix.Cmsghdr)(unsafe.Pointer(&oob[startLen]))
	h.Level = cmsgLevel
	h.Type = cmsgType
	h.SetLen(unix.CmsgLen(dataLen))
	dataOffset := startLen + unix.SizeofCmsghdr
	copy(oob[dataOffset:], data)
	return oob, oob[dataOffset:]
}
