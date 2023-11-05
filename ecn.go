package go_udp

import (
	"golang.org/x/sys/unix"
	"syscall"
)

type ECN uint8

const (
	ECNUnsupported ECN = iota
	ECNNon             // 00
	ECT1               // 01
	ECT0               // 10
	ECNCE              // 11
)

func (e ECN) ToHeaderBits() byte {
	//nolint:exhaustive // There are only 4 values.
	switch e {
	case ECNNon:
		return 0
	case ECT0:
		return 0b00000010
	case ECT1:
		return 0b00000001
	case ECNCE:
		return 0b00000011
	default:
		panic("ECN unsupported")
	}
}

func appendIPv4ECNMsg(oob []byte, val ECN) []byte {
	const dataLen = 4
	oob, cmsgData := appendCmsg(oob, syscall.IPPROTO_IP, unix.IP_TOS, dataLen)
	cmsgData[0] = val.ToHeaderBits()
	return oob
}

func appendIPv6ECNMsg(oob []byte, val ECN) []byte {
	const dataLen = 4
	oob, cmsgData := appendCmsg(oob, syscall.IPPROTO_IPV6, unix.IPV6_TCLASS, dataLen)
	cmsgData[0] = val.ToHeaderBits()
	return oob
}
