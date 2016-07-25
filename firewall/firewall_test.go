package firewall

import "testing"

func TestLocalNetworks(t *testing.T) {
	_, err := LocalAddrs()
	if err != nil {
		t.Error(err)
	}
}

func TestGrepV4Addr(t *testing.T) {
	addrs := []string{"127.0.0.1/16", "", "fe80::1/64", "192.168.0.1/24"}
	v4addrs := GrepV4Addr(addrs)
	if len(v4addrs) != 2 {
		t.Error("invalid number of v4addrs")
	}
	if v4addrs[0] != "127.0.0.1/16" {
		t.Error("first element is not 127.0.0.1")
	}
	if v4addrs[1] != "192.168.0.1/24" {
		t.Error("second element is not 192.168.0.1")
	}

}
