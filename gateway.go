package gateway

import (
	"log"
	"net"
	"net/netip"
)

var Version string = "v2.0.0"

type Network int

const (
	Udp Network = iota
	Tcp
)

func (d Network) String() string {
	return [...]string{"udp", "tcp"}[d]
}

type Gateway struct {
	Client  Direction
	Servers []Direction
}

var client connection
var servers []connection

type Direction struct {
	Net    Network
	Self   netip.AddrPort
	Remote netip.AddrPort
}

type connection struct {
	readBuffer []byte
	connUdp    *net.UDPConn
	//connTcp    *net.TCPConn
}

func (s Gateway) Run() error {
	for _, dir := range append(s.Servers, s.Client) {
		if !dir.Self.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Self.String()}
		}
		if !dir.Remote.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Remote.String()}
		}
	}

	switch s.Client.Net {
	case Udp:
		log.Printf("Подключение %s - %s\n", s.Client.Self, s.Client.Remote)
		conn1, err := net.DialUDP("udp",
			net.UDPAddrFromAddrPort(s.Client.Self),
			net.UDPAddrFromAddrPort(s.Client.Remote))
		if err != nil {
			log.Fatalf("Не удалось подключиться! %s - %s", s.Client.Self, s.Client.Remote)
			return &net.OpError{Op: "DialUDP", Net: "udp"}
		}
		client = connection{
			readBuffer: make([]byte, 300),
			connUdp:    conn1}
	case Tcp:

	}

	for _, dir := range s.Servers {
		switch s.Client.Net {
		case Udp:
			log.Printf("Подключение %s - %s\n", dir.Self, dir.Remote)
			conn1, err := net.DialUDP("udp",
				net.UDPAddrFromAddrPort(dir.Self),
				net.UDPAddrFromAddrPort(dir.Remote))
			if err != nil {
				log.Fatalf("Не удалось подключится! %s - %s", dir.Self, dir.Remote)
			} else {
				servers = append(servers, connection{
					readBuffer: make([]byte, 300),
					connUdp:    conn1})
			}
		case Tcp:

		}
	}

	for _, con1 := range servers {
		go s.transport(con1, []connection{client})
	}

	s.transport(client, servers)

	return nil
}

func (s Gateway) transport(src connection, dst []connection) {
	for {
		if src.connUdp != nil {
			n, _, _ := src.connUdp.ReadFromUDP(src.readBuffer)
			for _, c := range dst {
				if c.connUdp != nil {
					c.connUdp.Write(src.readBuffer[:n])
				}
			}
		}
	}
}
