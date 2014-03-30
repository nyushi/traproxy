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

func main() {
	var showVersion *bool = flag.Bool("V", false, "show version")
	var withDocker *bool = flag.Bool("with-docker", false, "edit iptables rule for docker")
	var withFirewall *bool = flag.Bool("with-fw", true, "edit iptables rule")
	var forceDstAddr *string = flag.String("dstaddr", "", "DEBUG force set to destination address")
	var proxyAddr *string = flag.String("proxyaddr", "", "proxy address")

	flag.Parse()

	if *showVersion {
		fmt.Println(traproxy.Version)
		os.Exit(0)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigc
		if *withFirewall {
			cmds := firewall.IptablesDel()
			if *withDocker {
				cmds = append(cmds, firewall.IptablesDockerDel()...)
			}
			for _, c := range cmds {
				execIptables(c)
			}
		}
		log.Fatal("finished")
	}()

	if *withFirewall {
		cmds := firewall.IptablesAdd()
		if *withDocker {
			cmds = append(cmds, firewall.IptablesDockerAdd()...)
		}
		for _, c := range cmds {
			execIptables(c)
		}
	}

	if *forceDstAddr != "" {
		d := Destination(*forceDstAddr)
		dst = &d
	}
	testo(*proxyAddr)
}

func execIptables(cmd firewall.IPTablesCommand) {
	out, err := cmd.Exec()
	if err != nil {
		log.Println(cmd, string(out))
		log.Fatal(err)
	}
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

func testo(proxyAddr string) {
	ln, err := net.Listen("tcp", ":10080")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("start server")
	for {
		client, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleClient(proxyAddr, client)
	}
}
