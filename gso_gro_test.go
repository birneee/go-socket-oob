package go_socket_oob

import (
	"github.com/stretchr/testify/require"
	"net"
	"net/netip"
	"testing"
)

func newGsoGroConn(t require.TestingT) *net.UDPConn {
	anyLocalAddr := netip.MustParseAddrPort("[::]:0")
	conn, err := net.ListenUDP("udp", net.UDPAddrFromAddrPort(anyLocalAddr))
	require.NoError(t, err)
	require.True(t, isGSOSupported(conn))
	require.True(t, isGROSupported(conn))
	err = enableGRO(conn)
	require.NoError(t, err)
	require.True(t, isGROEnabled(conn))
	return conn
}

func TestGsoGro(t *testing.T) {
	serverConn := newGsoGroConn(t)
	clientConn := newGsoGroConn(t)
	clientAddr := clientConn.LocalAddr().(*net.UDPAddr).AddrPort()
	go func() {
		buf := make([]byte, 10000)
		n, _, err := WriteGSO(serverConn, buf, 1000, clientAddr, nil)
		require.NoError(t, err)
		require.Equal(t, 10000, n)
	}()
	buf := make([]byte, 100000)
	result, err := ReadGRO(clientConn, buf, nil)
	require.NoError(t, err)
	require.Equal(t, 10000, len(result.FullBuf))
	require.Equal(t, 1000, result.SegmentSize)
}
