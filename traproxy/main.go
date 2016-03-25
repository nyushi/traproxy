package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

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
	showVersion := flag.Bool("V", false, "show version")
	withDocker := flag.Bool("with-docker", false, "DEPRECATED: edit iptables rule for docker.")
	withFirewall := flag.Bool("with-fw", true, "edit iptables rule")
	withFirewallNat := flag.Bool("with-fw-nat", true, "edit iptables rule with nat")
	excludeReservedAddrs := flag.Bool("exclude-reserved-addrs", false, "exclude reserved ip addresses")
	forceDstAddr := flag.String("dstaddr", "", "DEBUG force set to destination address")
	proxyAddr := flag.String("proxyaddr", "", "proxy address. '<host>:<port>'")
	var excludeAddrs excludeOptions
	flag.Var(&excludeAddrs, "exclude", "network addr to exclude")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s(%s)\n", traproxy.Version, traproxy.GitHash)
		os.Exit(0)
	}

	localAddrs, err := firewall.LocalAddrs()
	if err != nil {
		log.Fatal(err)
	}

	if *proxyAddr != "" {
		v := strings.SplitN(*proxyAddr, ":", 2)
		excludeAddrs = append(excludeAddrs, v[0])
	}
	excludeAddrs = append(excludeAddrs, firewall.GrepV4Addr(localAddrs)...)
	if *excludeReservedAddrs {
		excludeAddrs = append(excludeAddrs, firewall.ReservedV4Addrs()...)
	}
	redirectRules := firewall.GetRedirectRules(excludeAddrs)

	if *withFirewallNat {
		redirectRules = append(redirectRules, firewall.GetRedirectNATRules(excludeAddrs)...)
	}

	if *withDocker {

		log.Printf("waiting for %s", firewall.DockerIFName)
		if err := traproxy.WaitForCond(checkDockerInterface, time.Second*60); err != nil {
			msg := fmt.Sprintf("%s", err.Error())
			log.Fatal(msg)
		}
		log.Printf("%s detected", firewall.DockerIFName)
		redirectRules = append(
			redirectRules,
			firewall.GetRedirectDockerRules(excludeAddrs)...)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

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
		failed := false
		for _, r := range redirectRules {
			log.Println(r.GetCommandStr())
			err := r.Add()
			if err != nil {
				log.Println(err)
				failed = true
			}
		}
		if failed {
			log.Println("firewall setup failed. shutting down.")
			tearDown()
		}
	}

	if *forceDstAddr != "" {
		d := destination(*forceDstAddr)
		dst = &d
	}
	err = startServer(*proxyAddr)
	if err != nil {
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

func checkDockerInterface() (bool, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return false, err
	}
	for _, i := range interfaces {
		if i.Name == firewall.DockerIFName {
			return true, nil
		}
	}
	return false, nil
}
