package orgdst

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
)

const SO_ORIGINAL_DST = 80

func GetOriginalDst(c net.Conn) (string, error) {
	tcp, ok := c.(*net.TCPConn)
	if !ok {
		return "", errors.New("socket is not tcp")
	}
	file, err := tcp.File()
	if err != nil {
		return "", err
	}
	defer file.Close()
	fd := file.Fd()

	addr, err :=
		syscall.GetsockoptIPv6Mreq(
			int(fd),
			syscall.IPPROTO_IP,
			SO_ORIGINAL_DST)
	if err != nil {
		return "", err
	}

	ip := strings.Join([]string{
		itod(uint(addr.Multiaddr[4])),
		itod(uint(addr.Multiaddr[5])),
		itod(uint(addr.Multiaddr[6])),
		itod(uint(addr.Multiaddr[7])),
	}, ".")
	port := uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3])
	return fmt.Sprintf("%s:%d", ip, int(port)), nil
}
