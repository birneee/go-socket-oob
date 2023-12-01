package go_socket_oob

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"net/netip"
	"syscall"
	"unsafe"
)

type GroReadResult struct {
	FullBuf     []byte
	SegmentSize int
	OOB         []byte
	Flags       int
	RemoteAddr  netip.AddrPort
}

func appendUdpGenericReceiveOffloadMsg(oob []byte) (newOOB []byte) {
	const dataLen = 4
	oob, rawCmsgData := appendCmsg(oob, syscall.IPPROTO_UDP, unix.UDP_GRO, dataLen)
	*(*int)(unsafe.Pointer(&rawCmsgData[0])) = 1
	return oob
}

// ReadGRO uses generic receive offload
func ReadGRO(conn *net.UDPConn, b []byte, oob []byte) (GroReadResult, error) {
	oob = appendUdpGenericReceiveOffloadMsg(oob)
	n, oobn, flags, addr, err := conn.ReadMsgUDPAddrPort(b, oob)
	if err != nil {
		return GroReadResult{}, err
	}
	var segmentSize *int
	cmsgIter := iterateCmsgs(oob[:oobn])
	for cmsgIter.HasNext() {
		cmsg := cmsgIter.Next()
		if cmsg.Hdr.Level == unix.IPPROTO_UDP && cmsg.Hdr.Type == unix.UDP_GRO {
			segmentSize = (*int)(unsafe.Pointer(&cmsg.Data[0]))
		}
	}
	if segmentSize == nil {
		return GroReadResult{}, fmt.Errorf("gro failed")
	}
	return GroReadResult{
		FullBuf:     b[:n],
		SegmentSize: *segmentSize,
		OOB:         oob,
		Flags:       flags,
		RemoteAddr:  addr,
	}, nil
}

// isGROSupported tests if the kernel supports GRO.
// Sending with GRO might still fail later on, if the interface doesn't support it.
func isGROSupported(conn *net.UDPConn) bool {
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

// isGROEnabled tests if GSO is enabled on this socket.
// Sending with GRO might still fail later on, if the interface doesn't support it.
func isGROEnabled(conn *net.UDPConn) bool {
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

func enableGRO(conn *net.UDPConn) error {
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
