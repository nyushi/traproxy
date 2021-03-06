package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/nyushi/traproxy"
	"github.com/nyushi/traproxy/firewall"
	"github.com/nyushi/traproxy/orgdst"
)

type destination string

func (d *destination) Port() string {
	str := string(*d)
	_, port, _ := net.SplitHostPort(str)
	return port
}

var (
	dst *destination
)

type excludeOptions []string

func (e *excludeOptions) String() string {
	return fmt.Sprint(*e)
}
func (e *excludeOptions) Set(val string) error {
	for _, v := range strings.Split(val, ",") {
		*e = append(*e, v)
	}
	return nil
}

func main() {
	var withFirewallNat *bool
	showVersion := flag.Bool("V", false, "show version")
	withFirewall := flag.Bool("with-fw", true, "edit iptables rule")
	excludeReservedAddrs := flag.Bool("exclude-reserved-addrs", true, "exclude reserved ip addresses")
	forceDstAddr := flag.String("dstaddr", "", "DEBUG force set to destination address")
	proxyAddr := flag.String("proxyaddr", "", "proxy address. '<host>:<port>'")
	if runtime.GOOS == "linux" {
		withFirewallNat = flag.Bool("with-fw-nat", true, "edit iptables rule with nat")
	} else {
		b := true
		withFirewallNat = &b
	}
	var excludeAddrs excludeOptions
	flag.Var(&excludeAddrs, "exclude", "network addr to exclude")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s(%s)\n", traproxy.Version, traproxy.GitHash)
		os.Exit(0)
	}

	fwc := &firewall.Config{
		ProxyAddr:       proxyAddr,
		WithNat:         *withFirewallNat,
		ExcludeReserved: *excludeReservedAddrs,
		Excludes:        excludeAddrs,
	}
	if *withFirewall {
		switch runtime.GOOS {
		case "linux":
			fwc.FWType = firewall.FWIPTables
		case "darwin":
			fwc.FWType = firewall.FWPF
		}
	}
	fw := firewall.New(fwc)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	tearDown := func() {
		if err := fw.Teardown(); err != nil {
			log.Printf("error at teardown: %s", err)
		}
		log.Println("finished")
		os.Exit(0)
	}

	go func() {
		<-sigc
		tearDown()
	}()

	if *withFirewall {
		if err := fw.Setup(); err != nil {
			log.Printf("firewall setup failed. shutting down: %s", err)
			tearDown()
		}
	}

	if *forceDstAddr != "" {
		d := destination(*forceDstAddr)
		dst = &d
	}
	if err := startServer(*proxyAddr); err != nil {
		log.Println(err)
	}
	tearDown()
}

func getDst(c net.Conn) (destination, error) {
	if dst != nil {
		return *dst, nil
	}
	d, err := orgdst.GetOriginalDst(c)
	dst := destination(d)
	return dst, err
}

// StartProxy starts proxy process with client and proxy sockets
func StartProxy(client net.Conn, proxy net.Conn) {
	dst, err := getDst(client)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(dst)

	tbase := traproxy.TranslatorBase{
		Client: client,
		Proxy:  proxy,
		Dst:    string(dst),
	}

	var t traproxy.Translator
	if dst.Port() == "80" {
		t = &traproxy.HTTPTranslator{TranslatorBase: tbase}
	} else {
		t = &traproxy.HTTPSTranslator{TranslatorBase: tbase}
	}

	err = t.Start()
	if err != nil {
		panic(err)
	}
}

func handleClient(proxyAddr string, client net.Conn) {
	defer client.Close()
	defer func() {
		if e := recover(); e != nil {
			log.Printf("%s: %s", e, debug.Stack())
		}
	}()

	proxy, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		log.Printf("failed to connect proxy: %s\n", err.Error())
		return
	}
	defer proxy.Close()

	StartProxy(client, proxy)
}

func startServer(proxyAddr string) error {
	ln, err := net.Listen("tcp", ":10080")
	if err != nil {
		return err
	}
	log.Println("start server")
	for {
		client, err := ln.Accept()
		if err != nil {
			return err
		}

		go handleClient(proxyAddr, client)
	}
}
