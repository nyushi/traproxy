package orgdst

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// from https://github.com/nyushi/DIOCNATLOOK
const (
	pfOut       = 2
	diocnatlook = uintptr(3226747927)
)

// GetOriginalDst returns original destination of Conn
func GetOriginalDst(c net.Conn) (string, error) {
	nl := &pfiocNatlook{
		af: syscall.AF_INET,
	}
	remoteHost, remotePortStr, _ := net.SplitHostPort(c.RemoteAddr().String())
	remotePortInt, _ := strconv.Atoi(remotePortStr)
	localHost, localPortStr, _ := net.SplitHostPort(c.LocalAddr().String())
	localPortInt, _ := strconv.Atoi(localPortStr)

	raddr := net.ParseIP(remoteHost)
	laddr := net.ParseIP(localHost)
	(&nl.saddr).Set(raddr)
	(&nl.daddr).Set(laddr)
	(&nl.sxport).Set(uint16(remotePortInt))
	(&nl.dxport).Set(uint16(localPortInt))
	nl.proto = syscall.IPPROTO_TCP
	nl.direction = pfOut

	if err := lookup(nl); err != nil {
		return "", fmt.Errorf("failed to lookup: %s", err)
	}
	ip := strings.Join([]string{
		itod(uint(nl.rdaddr[0])),
		itod(uint(nl.rdaddr[1])),
		itod(uint(nl.rdaddr[2])),
		itod(uint(nl.rdaddr[3]))}, ".")
	return fmt.Sprintf("%s:%d", ip, nl.rdxport.Get()), nil
}

type pfAddr [16]byte

func (pa *pfAddr) Set(ip net.IP) {
	if ip.To4() != nil {
		// change alignment for in_addr
		ip = ip[12:16]
	}
	for i := 0; i < len(ip); i++ {
		pa[i] = ip[i]
	}
}

type pfPort [4]byte

func (pp *pfPort) Set(port uint16) {
	binary.BigEndian.PutUint16(pp[:], uint16(port))
}
func (pp pfPort) Get() (port uint16) {
	binary.Read(bytes.NewBuffer(pp[:]), binary.BigEndian, &port)
	return
}

type pfiocNatlook struct {
	saddr        pfAddr
	daddr        pfAddr
	rsaddr       pfAddr
	rdaddr       pfAddr
	sxport       pfPort
	dxport       pfPort
	rsxport      pfPort
	rdxport      pfPort
	af           uint8
	proto        uint8
	protoVariant uint8
	direction    uint8
}

func lookup(nl *pfiocNatlook) error {
	pfdev, err := syscall.Open("/dev/pf", syscall.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open /dev/pf: %s", err)
	}
	defer syscall.Close(pfdev)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(pfdev), diocnatlook, uintptr(unsafe.Pointer(nl)))
	if errno != 0 {
		return fmt.Errorf("ioctl error: %s", errno)
	}
	return nil
}
