package pkt

import (
	"syscall"
	"fmt"
	"os"
)

func Htons(n uint16) uint16 {
	var (
		high uint16 = n >> 8
		ret  uint16 = n<<8 + high
	)
	return ret
}

func Ioctl(fd, op, arg uintptr) error {
	_, _, ep := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(op), arg)
	if ep != 0 {
		return syscall.Errno(ep)
	}
	return nil
}

func Print(fd *os.File) {
	buf := make([]byte, 1024)
	numRead, _ := fd.Read(buf)
	fmt.Printf("%X\n", buf[:numRead])
}
