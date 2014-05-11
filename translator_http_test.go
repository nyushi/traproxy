package traproxy

import (
	"net"
	"testing"
	"time"
)

func getHTTPTranslator(network, endpoint string) (client, proxy *net.TCPConn, trans *HTTPTranslator, err error) {
	a, err := createSockets(network, endpoint)
	if err != nil {
		return nil, nil, nil, err
	}

	b, err := createSockets(network, endpoint)
	if err != nil {
		return nil, nil, nil, err
	}

	base := TranslatorBase{
		Client: a.B,
		Proxy:  b.B,
		Dst:    "example.com",
	}
	client, clientOk := a.A.(*net.TCPConn)
	proxy, proxyOk := b.A.(*net.TCPConn)
	if clientOk && proxyOk {
		return client, proxy, &HTTPTranslator{TranslatorBase: base}, nil
	}
	return nil, nil, &HTTPTranslator{TranslatorBase: base}, nil
}

func TestHTTPTranslatorStartSuccess(t *testing.T) {
	client, proxy, trans, err := getHTTPTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	go trans.Start()

	buf := make([]byte, 1024)
	client.Write([]byte("HEAD /test HTTP/1.0\r\nHost: localhost\r\n\r\n"))

	s, err := proxy.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	got := string(buf[0:s])
	expected := "HEAD http://localhost/test HTTP/1.0\r\nHost: localhost\r\n\r\n"
	if got != expected {
		t.Errorf("got=%s\nexpected=%s", got, expected)
	}
}

func TestHTTPTranslatorStartSuccessDoNothing(t *testing.T) {
	client, proxy, trans, err := getHTTPTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}
	go trans.Start()

	buf := make([]byte, 1024)
	client.Write([]byte("test"))

	proxy.SetDeadline(time.Now().Add(time.Millisecond))
	if _, err = proxy.Read(buf); err == nil {
		t.Fatal("not timeouted")
	}
}

func TestHTTPTranslatorStartNotTCP(t *testing.T) {
	_, _, trans, err := getHTTPTranslator("unix", "/tmp/traproxy_test")
	if err != nil {
		t.Error(err)
	}
	err = trans.Start()
	if err.Error() != "client socket is not tcp" {
		t.Error("socket check failed")
	}
}
