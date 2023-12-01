package go_socket_oob

import (
	"golang.org/x/sys/unix"
	"unsafe"
)

var CmsgHdrSpace int

func init() {
	CmsgHdrSpace = unix.CmsgSpace(0)
}

func appendCmsg(oob []byte, cmsgLevel int32, cmsgType int32, dataLen int) (newOob []byte, data []byte) {
	startLen := len(oob)
	oob = append(oob, make([]byte, unix.CmsgSpace(dataLen))...)
	h := (*unix.Cmsghdr)(unsafe.Pointer(&oob[startLen]))
	h.Level = cmsgLevel
	h.Type = cmsgType
	h.SetLen(unix.CmsgLen(dataLen))
	dataOffset := startLen + unix.CmsgSpace(0)
	return oob, oob[dataOffset : dataOffset+dataLen]
}

type Cmsg struct {
	Hdr  *unix.Cmsghdr
	Data []byte
}

type cmsgIterator struct {
	oob          []byte
	currentIndex int
}

func (g *cmsgIterator) HasNext() bool {
	return g.currentIndex < len(g.oob)
}

func (g *cmsgIterator) Next() Cmsg {
	hdr := (*unix.Cmsghdr)(unsafe.Pointer(&g.oob[g.currentIndex]))
	dataLen := int(hdr.Len) - CmsgHdrSpace
	space := unix.CmsgSpace(dataLen)
	data := g.oob[g.currentIndex+CmsgHdrSpace : g.currentIndex+CmsgHdrSpace+dataLen]
	g.currentIndex += space
	return Cmsg{
		Hdr:  hdr,
		Data: data,
	}
}

func iterateCmsgs(oob []byte) Iterator[Cmsg] {
	return &cmsgIterator{
		oob:          oob,
		currentIndex: 0,
	}
}
