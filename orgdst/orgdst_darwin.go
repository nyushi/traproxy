package orgdst

import (
	"errors"
	"net"
	"syscall"
)

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
	sa, err := syscall.Getsockname(int(file.Fd()))
	if err != nil {
		return "", err
	}

	var addr net.TCPAddr
	if v, ok := sa.(*syscall.SockaddrInet4); ok {
		addr = net.TCPAddr{IP: v.Addr[0:], Port: v.Port}
	} else if v, ok := sa.(*syscall.SockaddrInet6); ok {
		addr = net.TCPAddr{IP: v.Addr[0:], Port: v.Port, Zone: zoneToString(int(v.ZoneId))}
	} else {
		return "", errors.New("socket is not ipv4")
	}
	return addr.String(), nil
}
