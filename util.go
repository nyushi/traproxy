package traproxy

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type tcpconn interface {
	io.ReadWriter
	CloseRead() error
	CloseWrite() error
}

var pipeBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}

// Pipe starts bridging with two tcp connection
func Pipe(dst tcpconn, src tcpconn, f *func([]byte) []byte) error {
	defer src.CloseRead()
	defer dst.CloseWrite()

	rb := pipeBufPool.Get().([]byte)
	defer func() {
		pipeBufPool.Put(rb)
	}()

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

// WaitForCond wait until condition is true
func WaitForCond(cond func() (bool, error), timeout time.Duration) error {
	start := time.Now()
	for {
		ok, err := cond()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		if time.Now().Sub(start) > timeout {
			return fmt.Errorf("timed out")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
