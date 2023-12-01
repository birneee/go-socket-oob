package socket_oob

import (
	"github.com/stretchr/testify/require"
	"net"
	"testing"
	"time"
)

func BenchmarkGsoGro(b *testing.B) {
	bufSize := MaxGSOBufSize
	segmentSize := uint16(1500)

	serverConn := newGsoGroConn(b)
	clientConn := newGsoGroConn(b)
	clientAddr := clientConn.LocalAddr().(*net.UDPAddr).AddrPort()
	numBytesWritten := 0
	numBytesRead := 0
	stop := make(chan struct{})
	senderStopped := make(chan struct{})
	receiverStopped := make(chan struct{})
	go func() { // sender
		buf := make([]byte, bufSize)
	loop:
		for {
			n, _, err := WriteGSO(serverConn, buf, segmentSize, clientAddr, nil)
			require.NoError(b, err)
			select {
			case <-receiverStopped:
				serverConn.Close()
				close(senderStopped)
				break loop
			default: // continue
			}
			numBytesWritten += n
		}
	}()
	go func() { // receiver
		buf := make([]byte, bufSize)
	loop:
		for {
			segments, _, _, _, err := ReadGRO(clientConn, buf, nil)
			require.NoError(b, err)
			select {
			case <-stop:
				clientConn.Close()
				close(receiverStopped)
				break loop
			default: // continue
			}
			numBytesRead += len(segments.Buf)
		}
	}()
	b.ResetTimer()
	time.Sleep(100 * time.Duration(b.N))
	close(stop)
	<-receiverStopped
	b.StopTimer()
	b.ReportMetric(float64(numBytesWritten)/1e9, "GB_written")
	b.ReportMetric(float64(numBytesRead)/1e9, "GB_read")
	b.ReportMetric(float64(numBytesWritten)/1e9*8/b.Elapsed().Seconds(), "Gbps_written")
	b.ReportMetric(float64(numBytesRead)/1e9*8/b.Elapsed().Seconds(), "Gbps_read")
}
