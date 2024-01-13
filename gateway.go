package gateway

import (
	"log"
	"net"
	"net/netip"
	"regexp"
)

var Version string = "v0.3.4"

type Network int

const (
	Udp Network = iota
	Tcp
)

func (d Network) String() string {
	return [...]string{"udp", "tcp"}[d]
}

type gateway struct {
	Client  Direction
	Servers []Direction
}

type Gateway interface {
	Run() error
}

func New() Gateway {
	return &gateway{}
}

var client *connection
var servers []*connection

type Direction struct {
	Net    Network
	Self   netip.AddrPort
	Remote netip.AddrPort
}

type connection struct {
	readBuffer []byte
	connUdp    *net.UDPConn
	connTcp    *net.TCPConn
}

// Run Validate configuration and activate connections.
func (s *gateway) Run() error {
	for _, dir := range append(s.Servers, s.Client) {
		if !dir.Self.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Self.String()}
		}
		if !dir.Remote.IsValid() {
			return &net.AddrError{Err: "Address not valid", Addr: dir.Remote.String()}
		}
	}

	con1, err := s.connect(s.Client)
	if err == nil {
		client = con1
	}

	for _, dir := range s.Servers {
		con1, err := s.connect(dir)
		if err == nil {
			servers = append(servers, con1)
		}
	}

	for _, con1 := range servers {
		go s.transport(con1, []*connection{client})
	}

	s.transport(client, servers)

	return nil
}

// transport transmitting packets from the client to the servers
// and in the revers direction.
func (s *gateway) transport(src *connection, dst []*connection) {
	for {
		var n int
		if src.connUdp != nil {
			n, _, _ = src.connUdp.ReadFromUDP(src.readBuffer)
		}
		if src.connTcp != nil {
			n, _ = src.connTcp.Read(src.readBuffer)
		}
		for _, c := range dst {
			if c.connUdp != nil {
				c.connUdp.Write(src.readBuffer[:n])
			}
			if c.connTcp != nil {
				c.connTcp.Write(src.readBuffer[:n])
			}
		}
	}
}

// connect Connecting all directions.
func (s *gateway) connect(dir Direction) (*connection, error) {
	var err error
	switch dir.Net {
	case Udp:
		log.Printf("Подключение %s - %s\n", dir.Self, dir.Remote)
		conn1, err := net.DialUDP("udp", net.UDPAddrFromAddrPort(dir.Self), net.UDPAddrFromAddrPort(dir.Remote))
		if err != nil {
			log.Fatalf("Не удалось подключится! %s - %s", dir.Self, dir.Remote)
		} else {
			return &connection{
				readBuffer: make([]byte, 300),
				connUdp:    conn1}, nil
		}
	case Tcp:
		log.Printf("Подключение %s - %s\n", dir.Self, dir.Remote)
		conn2, err := net.DialTCP("tcp", net.TCPAddrFromAddrPort(dir.Self), net.TCPAddrFromAddrPort(dir.Remote))
		if err != nil {
			log.Fatalf("Не удалось подключится! %s - %s", dir.Self, dir.Remote)
		} else {
			return &connection{
				readBuffer: make([]byte, 300),
				connTcp:    conn2}, nil
		}
	}
	return nil, err
}

// SetConfig sets the configuration from lines like
// for master:
//
//	"127.0.0.1:40001-127.0.0.1:40002, udp"
//
// for slaves: [
//
//	"127.0.0.1:40011-127.0.0.1:43011, tcp, client",
//	"127.0.0.1:40012-127.0.0.1:43012, tcp, server",
//	].
func (g *gateway) SetConfig(client string, servers []string) error {

	var validDirection = regexp.MustCompile(`\s*([0-9.]+:[0-9]+)\s*-\s*([0-9.]+:[0-9]+)\s*,\s*(tcp|udp)\s*,*\s*(server|client)*\s*`)

	var dirs []Direction
	for _, strdir := range append([]string{client}, servers...) {
		if !validDirection.MatchString(strdir) {
			return &net.ParseError{}
		}
		res := validDirection.FindAllStringSubmatch(strdir, -1)

		var n Network = Udp
		if res[0][3] == "tcp" {
			n = Tcp
		}
		dir := Direction{
			Net:    n,
			Self:   netip.MustParseAddrPort(res[0][2]),
			Remote: netip.MustParseAddrPort(res[0][3]),
		}
		dirs = append(dirs, dir)
		log.Print(dir)
	}
	g.Client = dirs[0]
	g.Servers = dirs[1:]

	return nil
}
