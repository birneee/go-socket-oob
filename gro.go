package socket_oob

import (
	"golang.org/x/sys/unix"
	"net"
	"net/netip"
	"syscall"
	"unsafe"
)

func appendUdpGenericReceiveOffloadMsg(oob []byte) (newOOB []byte) {
	const dataLen = 4
	oob, rawCmsgData := appendCmsg(oob, syscall.IPPROTO_UDP, unix.UDP_GRO, dataLen)
	*(*int)(unsafe.Pointer(&rawCmsgData[0])) = 1
	return oob
}

// ReadGRO uses generic receive offload
func ReadGRO(conn *net.UDPConn, b []byte, oob []byte) (Segments, []byte, int, netip.AddrPort, error) {
	oob = appendUdpGenericReceiveOffloadMsg(oob)
	n, oobn, flags, addr, err := conn.ReadMsgUDPAddrPort(b, oob)
	if err != nil {
		return Segments{}, nil, 0, netip.AddrPort{}, err
	}
	segmentSize := n
	cmsgIter := iterateCmsgs(oob[:oobn])
	for cmsgIter.HasNext() {
		cmsg := cmsgIter.Next()
		if cmsg.Hdr.Level == unix.IPPROTO_UDP && cmsg.Hdr.Type == unix.UDP_GRO {
			segmentSize = *(*int)(unsafe.Pointer(&cmsg.Data[0]))
		}
	}
	return Segments{
		Buf:            b[:n],
		MaxSegmentSize: segmentSize,
	}, oob[:oobn], flags, addr, nil
}

// IsGROSupported tests if the kernel supports GRO.
// Sending with GRO might still fail later on, if the interface doesn't support it.
func IsGROSupported(conn *net.UDPConn) bool {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return false
	}
	var serr error
	if err := rawConn.Control(func(fd uintptr) {
		_, serr = unix.GetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_GRO)
	}); err != nil {
		return false
	}
	return serr == nil
}

// IsGROEnabled tests if GSO is enabled on this socket.
// Sending with GRO might still fail later on, if the interface doesn't support it.
func IsGROEnabled(conn *net.UDPConn) bool {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return false
	}
	var serr error
	var enabled = 0
	if err := rawConn.Control(func(fd uintptr) {
		enabled, serr = unix.GetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_GRO)
	}); err != nil {
		return false
	}
	return serr == nil && enabled != 0
}

func EnableGRO(conn *net.UDPConn) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	var serr error
	err = rawConn.Control(func(fd uintptr) {
		serr = unix.SetsockoptInt(int(fd), unix.IPPROTO_UDP, unix.UDP_GRO, 1)
	})
	if serr != nil {
		return serr
	}
	if err != nil {
		return err
	}
	return nil
}
