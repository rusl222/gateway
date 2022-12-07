package gateway

import (
	"fmt"
	"log"
	"net"
	"net/netip"
)

// --- class v1.0.0
type Gateway struct {
	Directions  []Direction
	connections []Connection
}

type Direction struct {
	Self   netip.AddrPort
	Remote netip.AddrPort
}

type Connection struct {
	readBuffer []byte
	conn       *net.UDPConn
}

func (s Gateway) Run() error {
	for _, dir := range s.Directions {
		if !dir.Self.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Self.String()}
		}
		if !dir.Remote.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Remote.String()}
		}

		log.Printf("connecting %s - %s\n", dir.Self, dir.Remote)
		conn1, err := net.DialUDP("udp",
			net.UDPAddrFromAddrPort(dir.Self),
			net.UDPAddrFromAddrPort(dir.Remote))
		if err != nil {
			fmt.Printf("Не удалось создать подключение %s - %s", dir.Self, dir.Remote)
			return &net.OpError{Op: "DialUDP", Net: "udp"}
		}

		s.connections = append(s.connections, Connection{
			readBuffer: make([]byte, 300),
			conn:       conn1})
	}

	for i := 1; i < len(s.connections); i++ {
		go s.transport(s.connections[i], []Connection{s.connections[0]})
	}

	s.transport(s.connections[0], s.connections[1:])

	return nil
}

func (s Gateway) transport(src Connection, dst []Connection) {
	for {
		n, _, _ := src.conn.ReadFromUDP(src.readBuffer)
		for _, c := range dst {
			c.conn.Write(src.readBuffer[:n])
		}
	}
}
