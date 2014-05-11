package traproxy

import (
	"errors"
	"log"
	"net"
	"runtime/debug"
)

// Translator is the interface that wraps the proxy translation
type Translator interface {
	Start() error
}

// TranslatorBase contains client/proxy socket and destination
type TranslatorBase struct {
	Client net.Conn
	Proxy  net.Conn
	Dst    string
}

// CheckSockets check Conn and returns TCPConn
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

// HandlePanic is utility for recovering panic in goroutine
func (t *TranslatorBase) HandlePanic() {
	if e := recover(); e != nil {
		log.Printf("%s: %s", e, debug.Stack())
	}
}
