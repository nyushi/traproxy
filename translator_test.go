package traproxy

import (
	"bytes"
	"log"
	"net"
	"os"
	"testing"
)

type Sockets struct {
	A net.Conn
	B net.Conn
}

func createSockets(network, endpoint string) (*Sockets, error) {
	ln, err := net.Listen(network, endpoint)
	if err != nil {
		return nil, err
	}

	c := make(chan bool)

	var client, server net.Conn
	go func() {
		c <- true
		server, _ = ln.Accept()
		ln.Close()
		c <- true
	}()
	<-c
	client, err = net.Dial(network, endpoint)
	if err != nil {
		return nil, err
	}
	<-c
	return &Sockets{
		A: server,
		B: client,
	}, nil
}

func TestTranslatorBaseCheckSocketsSuccess(t *testing.T) {
	s, err := createSockets("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	trans := TranslatorBase{
		Client: s.A,
		Proxy:  s.B,
		Dst:    "dst",
	}
	a, b, err := trans.CheckSockets()
	if err != nil {
		t.Error(err)
	}
	if a == nil || b == nil {
		t.Error("client is not TCPSock")
	}
}

func TestTranslatorBaseCheckSocketsClientIsUnix(t *testing.T) {
	c, err := createSockets("unix", "/tmp/traproxy_test")
	if err != nil {
		t.Fatal(err)
	}
	p, err := createSockets("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	trans := TranslatorBase{
		Client: c.A,
		Proxy:  p.A,
		Dst:    "dst",
	}
	a, b, err := trans.CheckSockets()
	if err == nil {
		t.Error("error not returned")
	}
	if a != nil || b != nil {
		t.Error("return value is not nil")
	}
}

func TestTranslatorBaseCheckSocketsProxyIsUnix(t *testing.T) {
	c, err := createSockets("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	p, err := createSockets("unix", "/tmp/traproxy_test")
	if err != nil {
		t.Fatal(err)
	}

	trans := TranslatorBase{
		Client: c.A,
		Proxy:  p.A,
		Dst:    "dst",
	}
	a, b, err := trans.CheckSockets()
	if err == nil {
		t.Error("error not returned")
	}
	if a != nil || b != nil {
		t.Error("return value is not nil")
	}
}

func TestTranslatorBaseHandlePanic(t *testing.T) {
	trans := &TranslatorBase{}
	buf := bytes.NewBuffer([]byte{})
	log.SetOutput(buf)
	c := make(chan bool)

	go func() {
		defer func() {
			c <- true
		}()
		defer trans.HandlePanic()
		panic("dummy")
	}()
	<-c
	if !bytes.Contains(buf.Bytes(), []byte("dummy")) {
		t.Error("recover failed")
	}
	log.SetOutput(os.Stdout)
}
