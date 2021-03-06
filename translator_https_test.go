package traproxy

import (
	"net"
	"strings"
	"testing"
)

func getHTTPSTranslator(network, endpoint string) (client, proxy *net.TCPConn, trans *HTTPSTranslator, err error) {
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
		return client, proxy, &HTTPSTranslator{base}, nil
	}
	return nil, nil, &HTTPSTranslator{base}, nil
}

func TestHTTPSTranslatorStartSuccess(t *testing.T) {
	client, proxy, trans, err := getHTTPSTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Error(err)
	}
	go trans.Start()

	buf := make([]byte, 1024)
	s, err := proxy.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := "CONNECT example.com HTTP/1.1\r\n\r\n"
	actual := string(buf[0:s])
	if actual != expected {
		t.Errorf("connect request error\nact='%s'\nexp='%s'", actual, expected)
	}

	proxy.Write([]byte("HTTP/1.1 200 Connection established\r\n"))

	client.Write([]byte("this is data"))
	s, err = proxy.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	expected = "this is data"
	actual = string(buf[0:s])
	if actual != expected {
		t.Errorf("write data error\nact=%s\nexp=%s", actual, expected)
	}
}

func TestHTTPSTranslatorStartResponseError(t *testing.T) {
	c := make(chan error)
	_, proxy, trans, err := getHTTPSTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Error(err)
	}
	go func() {
		c <- trans.Start()
	}()

	buf := make([]byte, 1024)
	s, err := proxy.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := "CONNECT example.com HTTP/1.1\r\n\r\n"
	actual := string(buf[0:s])
	if actual != expected {
		t.Errorf("connect request error\nact='%s'\nexp='%s'", actual, expected)
	}

	proxy.Write([]byte("this is invalid response"))
	err = <-c
	if err.Error() != "error response at CONNECT request: this is invalid response" {
		t.Error("response error not returned")
	}
}

func TestHTTPSTranslatorStartWriteError(t *testing.T) {
	_, _, trans, err := getHTTPSTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Error(err)
	}
	proxy := trans.Proxy.(*net.TCPConn)
	proxy.CloseWrite()
	err = trans.Start()
	if !strings.Contains(err.Error(), "failed to write at CONNECT:") {
		t.Error("write error not returned")
	}
}

func TestHTTPSTranslatorStartReadError(t *testing.T) {
	_, _, trans, err := getHTTPSTranslator("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Error(err)
	}
	proxy := trans.Proxy.(*net.TCPConn)
	proxy.CloseRead()
	err = trans.Start()
	if err.Error() != "failed to read at CONNECT: EOF" {
		t.Error("write error not returned")
	}
}

func TestHTTPSTranslatorStartNotTCP(t *testing.T) {
	_, _, trans, err := getHTTPSTranslator("unix", "/tmp/traproxy_test")
	if err != nil {
		t.Error(err)
	}
	err = trans.Start()
	if err.Error() != "client socket is not tcp" {
		t.Error("socket check failed")
	}
}
