package traproxy

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func getSockets() (a1, a2, b1, b2 *net.TCPConn, e error) {
	ln, err := net.Listen("tcp", ":60606")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	sockChan := make(chan *net.TCPConn, 2)
	errChan := make(chan error)
	go func() {
		for i := 0; i < 2; i++ {
			c, err := ln.Accept()
			if err != nil {
				errChan <- err
			}
			conn, _ := c.(*net.TCPConn)
			sockChan <- conn
		}
	}()

	c, err := net.Dial("tcp", "localhost:60606")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	a1 = c.(*net.TCPConn)

	c, err = net.Dial("tcp", "localhost:60606")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	b1 = c.(*net.TCPConn)
	select {
	case a2 = <-sockChan:
	case err := <-errChan:
		return nil, nil, nil, nil, err
	}
	select {
	case b2 = <-sockChan:
	case err := <-errChan:
		return nil, nil, nil, nil, err
	}
	return a1, a2, b1, b2, nil
}

func TestPipe(t *testing.T) {
	a1, a2, b1, b2, err := getSockets()
	if err != nil {
		t.Error(err)
	}

	go Pipe(b2, a2, nil)

	wb := []byte("123")
	rb := make([]byte, 1024)
	a1.Write(wb)
	size, err := b1.Read(rb)
	if err != nil {
		t.Error(err)
	}

	if size != 3 {
		t.Errorf("read size error: expected=%d, got=%d", 3, size)
	}
	if string(rb[:size]) != "123" {
		t.Errorf("read data error: expected='123', got=%s", string(rb[:size]))
	}
}

func TestWaitForCodn(t *testing.T) {
	start := time.Now()
	WaitForCond(func() (bool, error) { return true, nil }, time.Second)
	if time.Now().Sub(start) > time.Second {
		t.Errorf("not returned soon")
	}

	val := true
	start = time.Now()
	WaitForCond(func() (bool, error) {
		// first call is false
		// second call is true
		val = !val
		return val, nil
	}, time.Second)
	if time.Now().Sub(start) > time.Second {
		t.Errorf("not returned soon")
	}

	start = time.Now()
	WaitForCond(func() (bool, error) { return false, nil }, time.Second)
	if time.Now().Sub(start) < time.Second {
		t.Errorf("returned soon")
	}

	start = time.Now()
	err := WaitForCond(func() (bool, error) { return false, fmt.Errorf("err") }, time.Second)
	if err == nil {
		t.Errorf("error not returned")
	}
}
