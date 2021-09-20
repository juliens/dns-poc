package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"
)

func main() {
	cache = map[string]net.IP{}
	// a.root-servers.net
	server := flag.String("root-server", "198.41.0.4", "root server ip to use")
	listen := flag.String("listen", ":8053", "address to listen on")
	flag.Parse()

	root = net.ParseIP(*server)

	log.Printf("Listen on %q", *listen)
	err := dns.ListenAndServe(*listen, "udp", dns.HandlerFunc(func(writer dns.ResponseWriter, msg *dns.Msg) {
		for _, q := range msg.Question {
			if q.Qtype != dns.TypeA {
				continue
			}
			ip, msgResponse, err := findDNS(q.Name, root, "")
			if err != nil {
				writer.WriteMsg(msgResponse)
				return
			}
			
			rr, err := dns.NewRR(fmt.Sprintf("%s %d IN A %s", q.Name, 60, ip))
			if err != nil {
				continue
			}

			m := new(dns.Msg)
			m.SetReply(msg)
			m.Answer = append(m.Answer, rr)
			err = writer.WriteMsg(m)
			if err != nil {
				log.Fatal(err)
			}
		}
	}))
	if err != nil {
		log.Fatal(err)
	}
}

var root net.IP
var cache map[string]net.IP

func findDNS(name string, dnsServer net.IP, logPrefix string) (net.IP, *dns.Msg, error) {
	log.Printf(logPrefix + "ASK %s, %s", name, dnsServer)
	if ip, ok := cache[name]; ok {
		log.Printf(logPrefix + "FOUND IN CACHE %s %s", name, ip)
		return ip, nil, nil
	}

	if name[len(name)-1:] != "." {
		name += "."
	}
	m := new(dns.Msg)
	m.SetQuestion(name, dns.TypeA)

	c := new(dns.Client)
	conn, err := c.Dial(dnsServer.String()+":53")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial: %w", err)
	}
	if conn == nil {
		return nil, nil, errors.New("conn is nil")
	}
	err = conn.WriteMsg(m)
	if err != nil {
		return nil, nil, err
	}

	msg, err := conn.ReadMsg()
	if err != nil {
		return nil, nil, err
	}
	if msg.Rcode != dns.RcodeSuccess {
		return nil, msg, fmt.Errorf("error while resolve: %s", dns.RcodeToString[msg.Rcode])
	}


	for _, rr := range msg.Extra {
		switch x := rr.(type) {
		case *dns.A:
			cache[x.Hdr.Name] = x.A
		}
	}

	for _, rr := range msg.Answer {
		switch x := rr.(type) {
		case *dns.A:
			log.Printf(logPrefix + "FOUND %s %s", name, x.A)
			return x.A, nil, nil
		case *dns.CNAME:
			log.Printf(logPrefix + "FIND CNAME %s", x.Target)
			return findDNS(x.Target, root, logPrefix)
		default:
			log.Printf(logPrefix + "Ignore %s", x)
		}
	}

	for _, rr := range msg.Ns {
		switch x := rr.(type) {
		case *dns.NS:
			log.Printf(logPrefix + "NS %s", x.Ns)
			ip, msgResponse, err := findDNS(x.Ns, root, logPrefix + "----")
			if err != nil {
				return nil, msgResponse, err
			}
			return findDNS(name, ip, logPrefix)
		case *dns.A:
			return x.A, nil, nil
		case *dns.CNAME:
			return findDNS(x.Target, root, logPrefix)
		default:
			log.Printf(logPrefix + "Ignore %s", x)
		}
	}

	return nil, nil, nil
}
