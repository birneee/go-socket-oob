package socket_oob

import (
	"golang.org/x/sys/unix"
	"math"
	"net"
	"net/netip"
	"syscall"
	"unsafe"
)

const MaxGSOBufSize = math.MaxUint16 - 8

func appendUDPSegmentSizeMsg(oob []byte, size uint16) []byte {
	const dataLen = 2 // payload is a uint16
	oob, cmsgData := appendCmsg(oob, syscall.IPPROTO_UDP, unix.UDP_SEGMENT, dataLen)
	*(*uint16)(unsafe.Pointer(&cmsgData[0])) = size
	return oob
}

// WriteGSO uses generic segmentation offload
func WriteGSO(conn *net.UDPConn, b []byte, size uint16, addr netip.AddrPort, oob []byte) (n int, oobn int, err error) {
	oob = appendUDPSegmentSizeMsg(oob, size)
	return conn.WriteMsgUDPAddrPort(b, oob, addr)
}

// IsGSOSupported tests if the kernel supports GSO.
// Sending with GSO might still fail later on, if the interface doesn't support it.
func IsGSOSupported(conn *net.UDPConn) bool {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return false
	}
	var serr error
	if err := rawConn.Control(func(fd uintptr) {
		_, serr = unix.GetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_SEGMENT)
	}); err != nil {
		return false
	}
	return serr == nil
}
