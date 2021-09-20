# dns-poc

A small poc to resolve ip from domain name without cache by using authoritative name servers.

```
go run main.go
```

```
drill @127.0.0.1 -p 8053 my.example.com
```

## Resources:

- [https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml](https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml)
- [https://www.iana.org/domains/root/servers](https://www.iana.org/domains/root/servers)
- [https://www.iana.org/domains/root](https://www.iana.org/domains/root)
- https://cisco.goffinet.org/ccna/services-infrastructure/protocole-resolution-noms-dns/