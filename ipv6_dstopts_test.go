package go_socket_oob

import (
	"github.com/stretchr/testify/assert"
	"net"
	"net/netip"
	"testing"
)

// test might require root privileges
func TestSendSingleBlock(t *testing.T) {
	udpConn, err := net.ListenUDP("udp6", net.UDPAddrFromAddrPort(netip.AddrPortFrom(netip.IPv6Loopback(), 0)))
	assert.NoError(t, err)
	localAddr := udpConn.LocalAddr().(*net.UDPAddr).AddrPort()
	assert.NoError(t, err)
	buf := make([]byte, 1484)
	oob := make([]byte, 0)
	dstOptsData := appendPadDestOpt(nil, 6)
	oob, dstOptsDataBuf, err := appendDestOpts(oob, len(dstOptsData))
	assert.NoError(t, err)
	copy(dstOptsDataBuf, dstOptsData)
	n, oobn, err := udpConn.WriteMsgUDPAddrPort(buf, oob, localAddr)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, len(oob), oobn)
	assert.NoError(t, err)
}

// test might require root privileges
func TestSendTwoBlocks(t *testing.T) {
	udpConn, err := net.ListenUDP("udp6", net.UDPAddrFromAddrPort(netip.AddrPortFrom(netip.IPv6Loopback(), 0)))
	assert.NoError(t, err)
	localAddr := udpConn.LocalAddr().(*net.UDPAddr).AddrPort()
	assert.NoError(t, err)
	buf := make([]byte, 1476)
	oob := make([]byte, 0)
	dstOptsData := appendPadDestOpt(nil, 14)
	oob, dstOptsDataBuf, err := appendDestOpts(oob, len(dstOptsData))
	assert.NoError(t, err)
	copy(dstOptsDataBuf, dstOptsData)
	n, oobn, err := udpConn.WriteMsgUDPAddrPort(buf, oob, localAddr)
	assert.Equal(t, len(buf), n)
	assert.Equal(t, len(oob), oobn)
	assert.NoError(t, err)
}
