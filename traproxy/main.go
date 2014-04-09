package main

import (
	"flag"
	"fmt"
	"github.com/nyushi/traproxy"
	"github.com/nyushi/traproxy/firewall"
	"github.com/nyushi/traproxy/orgdst"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
)

type Destination string

func (d *Destination) Port() string {
	str := string(*d)
	_, port, _ := net.SplitHostPort(str)
	return port
}

var (
	dst *Destination = nil
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
	var showVersion *bool = flag.Bool("V", false, "show version")
	var withDocker *bool = flag.Bool("with-docker", false, "edit iptables rule for docker")
	var withFirewall *bool = flag.Bool("with-fw", true, "edit iptables rule")
	var forceDstAddr *string = flag.String("dstaddr", "", "DEBUG force set to destination address")
	var proxyAddr *string = flag.String("proxyaddr", "", "proxy address")
	var excludeAddrs excludeOptions
	flag.Var(&excludeAddrs, "exclude", "network addr to exclude")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s(%s)\n", traproxy.Version, traproxy.GitHash)
		os.Exit(0)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	localAddrs, err := firewall.LocalAddrs()
	if err != nil {
		log.Fatal(err)
	}

	excludeAddrs = append(excludeAddrs, firewall.GrepV4Addr(localAddrs)...)
	redirectRules := firewall.GetRedirectRules(excludeAddrs)
	if *withDocker {
		redirectRules = append(
			redirectRules,
			firewall.GetRedirectDockerRules(excludeAddrs)...)
	}

	tearDown := func() {
		if *withFirewall {
			for _, r := range redirectRules {
				log.Println(r.GetCommandStr())
				r.Del()
			}
		}
		log.Println("finished")
		os.Exit(0)
	}

	go func() {
		<-sigc
		tearDown()
	}()

	if *withFirewall {
		for _, r := range redirectRules {
			log.Println(r.GetCommandStr())
			r.Add()
		}
	}

	if *forceDstAddr != "" {
		d := Destination(*forceDstAddr)
		dst = &d
	}
	err = startServer(*proxyAddr)
	if err != nil {
		log.Println(err)
	}
	tearDown()
}

func getDst(c net.Conn) (Destination, error) {
	if dst != nil {
		return *dst, nil
	}
	d, err := orgdst.GetOriginalDst(c)
	dst := Destination(d)
	return dst, err
}

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
