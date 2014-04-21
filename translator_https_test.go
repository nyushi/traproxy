package traproxy

import (
	"net"
	"testing"
)

func TestHTTPSTranslator(t *testing.T) {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		t.Fatal(err)
	}

	c := make(chan bool)
	var serv1, serv2 net.Conn

	go func() {
		serv1, err = ln.Accept()
		if err != nil {
			t.Fatal(err)
		}

		serv2, err = ln.Accept()
		if err != nil {
			t.Fatal(err)
		}
		c <- true
	}()

	client1, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	client2, err := net.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Fatal(err)
	}

	<-c

	base := TranslatorBase{
		Client: client1,
		Proxy:  client2,
		Dst:    "example.com",
	}
	trans := &HTTPSTranslator{base}
	go trans.Start()

	buf := make([]byte, 1024)
	s, err := serv2.Read(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := "CONNECT example.com HTTP/1.1\r\n\r\n"
	actual := string(buf[0:s])
	if actual != expected {
		t.Errorf("connect request error\nact='%s'\nexp='%s'", actual, expected)
	}

	serv2.Write([]byte("HTTP/1.1 200 Connection established\r\n"))

	serv1.Write([]byte("this is data"))
	s, err = serv2.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	expected = "this is data"
	actual = string(buf[0:s])
	if actual != expected {
		t.Errorf("write data error\nact=%s\nexp=%s", actual, expected)
	}
}
