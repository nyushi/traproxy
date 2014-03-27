package traproxy

import (
	"errors"
	"log"
	"net"
	"runtime/debug"
)

type Translator interface {
	Start() error
}

type TranslatorBase struct {
	Client net.Conn
	Proxy  net.Conn
	Dst    string
}

func (t *TranslatorBase) CheckSockets() (*net.TCPConn, *net.TCPConn, error) {
	client, ok := t.Client.(*net.TCPConn)
	if !ok {
		return nil, nil, errors.New("client socket is not tcp")
	}
	proxy, ok := t.Proxy.(*net.TCPConn)
	if !ok {
		return nil, nil, errors.New("proxy socket is not tcp")
	}
	return client, proxy, nil
}

func (t *TranslatorBase) HandlePanic() {
	if e := recover(); e != nil {
		log.Printf("%s: %s", e, debug.Stack())
	}
}
