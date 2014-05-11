package traproxy

import (
	"net"
)

// Pipe starts bridging with two tcp connection
func Pipe(dst *net.TCPConn, src *net.TCPConn, f *func([]byte) []byte) error {
	defer src.CloseRead()
	defer dst.CloseWrite()

	rb := make([]byte, 4096)

	for {
		rsize, err := src.Read(rb)
		if err != nil {
			if isRecoverable(err) {
				continue
			}
			return err
		}

		var wb []byte
		if f != nil {
			wb = (*f)(rb[:rsize])
		} else {
			wb = rb[:rsize]
		}
		wWrote := 0
		wTotal := len(wb)
		for wWrote != wTotal {
			wSize, err := dst.Write(wb[wWrote:])
			wWrote += wSize
			if err != nil {
				if isRecoverable(err) {
					continue
				}
				return err
			}
		}
	}
}

func isRecoverable(e error) bool {
	ne, ok := e.(net.Error)
	if !ok {
		return false
	}
	return ne.Temporary()
}
