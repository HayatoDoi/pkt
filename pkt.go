package pkt

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

const (
	PF_PACKET = 17 // taken from /usr/include/x86_64-linux-gnu/bits/socket.h
)

type packetMreq struct {
	mrIfindex int32
	mrType    uint16
	mrAlen    uint16
	mrAddress [8]uint8
}

func htons(host uint16) uint16 {
	return (host&0xff)<<8 | (host >> 8)
}

type PFConn struct {
	fd   int
	intf *net.Interface
}

func (c *PFConn) Read(b []byte) (int, error) {
	for {
		n, from, err := syscall.Recvfrom(c.fd, b, 0)
		if err != nil {
			return 0, err
		}
		sa, _ := from.(*syscall.SockaddrLinklayer)
		if sa.Pkttype != syscall.PACKET_OUTGOING {
			return n, nil
		}
	}
}

func (c *PFConn) Write(b []byte) error {
	return syscall.Sendto(c.fd, b, 0, &syscall.SockaddrLinklayer{
		Ifindex:  c.intf.Index,
		Protocol: htons(syscall.ETH_P_ALL),
	})
}

func (c *PFConn) Close() error {
	_, _, e := syscall.Syscall(syscall.SYS_CLOSE, uintptr(c.fd), 0, 0)
	if e > 0 {
		return e
	}
	return nil
}

func (c *PFConn) String() string {
	return fmt.Sprintf("Conn(fd: %d, intf: %v)", c.fd, c.intf)
}

func NewPFConn(ifname string) (*PFConn, error) {
	intf, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.Socket(PF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))
	if err != nil {
		return nil, err
	}
	mreq := packetMreq{
		mrIfindex: int32(intf.Index),
		mrType:    syscall.PACKET_MR_PROMISC,
	}
	if _, _, e := syscall.Syscall6(syscall.SYS_SETSOCKOPT, uintptr(fd),
		uintptr(syscall.SOL_PACKET), uintptr(syscall.PACKET_ADD_MEMBERSHIP),
		uintptr(unsafe.Pointer(&mreq)), unsafe.Sizeof(mreq), 0); e > 0 {
		return nil, e
	}
	sll := syscall.RawSockaddrLinklayer{
		Family:   PF_PACKET,
		Protocol: htons(syscall.ETH_P_ALL),
		Ifindex:  int32(intf.Index),
	}
	if _, _, e := syscall.Syscall(syscall.SYS_BIND, uintptr(fd),
		uintptr(unsafe.Pointer(&sll)), unsafe.Sizeof(sll)); e > 0 {
		return nil, e
	}
	return &PFConn{
		fd:   fd,
		intf: intf,
	}, nil
}
