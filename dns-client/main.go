package main

import (
	"github.com/miekg/dns"
	"net"
	"os"
	"log"
	"fmt"
)

func main() {
	//config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	config := new(dns.ClientConfig)
	config.Servers = []string{"151.92.251.24"}
	config.Port = "53"
	config.Ndots = 1
	config.Timeout = 5
	config.Attempts = 2

	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("vldap.fg.local"), dns.TypeA)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil {
		log.Fatalf("*** error: %s\n", err.Error())
	}

	if r.Rcode != dns.RcodeSuccess {
		log.Fatalf(" *** invalid answer name %s after MX query for %s\n", os.Args[1], os.Args[1])
	}
	// Stuff must be in the answer section
	for _, a := range r.Answer {
		fmt.Printf("%v\n", a)
	}
}