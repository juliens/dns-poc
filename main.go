package main

import (
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

func main() {
	cache = map[string]net.IP{}
	// a.root-servers.net
	root = net.ParseIP("198.41.0.4")

	dns.ListenAndServe("127.0.0.1:8053", "udp", dns.HandlerFunc(func(writer dns.ResponseWriter, msg *dns.Msg) {
		for _, q := range msg.Question {
			if q.Qtype != dns.TypeA {
				continue
			}
			ip := findDNS(q.Name, root)
			fmt.Println(ip)
			rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", q.Name, 60, ip))
			if err != nil {
				continue
			}

			m := new(dns.Msg)
			m.SetReply(msg)
			m.Answer = append(m.Answer, rr)
			writer.WriteMsg(m)
		}
	}) )
}

var root net.IP
var cache map[string]net.IP

func findDNS(name string, dnsServer net.IP) net.IP {
	fmt.Println("ASK", name, dnsServer)
	if ip, ok := cache[name]; ok {
		fmt.Println("FOUND IN CACHE", name, ip)
		return ip
	}

	if name[len(name)-1:] != "." {
		name += "."
	}
 	m := new(dns.Msg)
	m.SetQuestion(name, dns.TypeA)

	c := new(dns.Client)
	conn, err := c.Dial(dnsServer.String()+":53")
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	if conn == nil {
		log.Fatalf("conn is nil")
	}
	err = conn.WriteMsg(m)
	if err != nil {
		log.Fatal(m, err)
	}

	msg, err := conn.ReadMsg()
	// fmt.Println(msg)

	for _, rr := range msg.Extra {
		switch x := rr.(type) {
			case *dns.A:
				// fmt.Println("Add in cache", x.Hdr.Name, x.A)
				cache[x.Hdr.Name] = x.A
		}
	}

	for _, rr := range msg.Answer {
		switch x := rr.(type) {
		case *dns.A:
			fmt.Println("FOUND", name, x.A)
			return x.A
		case *dns.CNAME:
			fmt.Println("FIND CNAME", x.Target)
			return findDNS(x.Target, root)
		default:
			log.Println("Ignore", x)
		}
	}

	for _, rr := range msg.Ns {
		switch x := rr.(type) {
		case *dns.NS:
			fmt.Println("NS", x.Ns)
			return findDNS(name, findDNS(x.Ns, root))
		case *dns.A:
			// fmt.Println("FOUND", name, x.A)
			return x.A
		case *dns.CNAME:
			return findDNS(x.Target, root)
		default:
			log.Println("Ignore", x)
		}
	}

	return nil
}