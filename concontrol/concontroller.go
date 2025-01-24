package concontrol

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"
)

type ConControllerConfig struct {
	Network string
	Role    string
	Self    netip.AddrPort
	Remote  netip.AddrPort
	Name    string
}

// ConController is a controller for a connection.
type ConController struct {
	Conn net.Conn
	conf ConControllerConfig
}

func New(conf ConControllerConfig) (*ConController, error) {
	if conf.Network != "tcp" && conf.Network != "udp" {
		return nil, &net.AddrError{Err: "Network not supported", Addr: conf.Network}
	}
	if !conf.Self.IsValid() {
		return nil, &net.AddrError{Err: "Address not valid", Addr: conf.Self.String()}
	}
	if !conf.Remote.IsValid() {
		return nil, &net.AddrError{Err: "Address not valid", Addr: conf.Remote.String()}
	}
	return &ConController{
		conf: conf,
	}, nil
}

func (c *ConController) Reconnect() {
	c.Conn = nil
}

func (c *ConController) Name() string {
	return c.conf.Name
}

func (c *ConController) Run(ctx context.Context) {
	var list net.Listener
	var err error
	fmt.Printf("[INF] %s: Running\n", c.conf.Name)
	for {
		select {
		case <-ctx.Done():
			c.Conn.Close()
			return
		default:
			if c.Conn == nil {
				switch c.conf.Network + c.conf.Role {
				// client role
				case "tcpm":
					addrSelf := net.TCPAddrFromAddrPort(c.conf.Self)
					addrRemote := net.TCPAddrFromAddrPort(c.conf.Remote)

					conn, err := net.DialTCP("tcp", addrSelf, addrRemote)
					if err != nil {
						fmt.Printf("[ERR] %s: %v -> sleep 10 sec\n", c.conf.Name, err)
						time.Sleep(10 * time.Second)
						continue
					}
					c.Conn = conn

				//server role
				case "tcps", "tcpp":

					if list == nil {
						list, err = net.Listen("tcp", c.conf.Self.String())
						if err != nil {
							fmt.Printf("[ERR] %s: %v -> sleep 10 sec\n", c.conf.Name, err)
							time.Sleep(10 * time.Second)
							continue
						}
					}
					conn, err := list.Accept()
					if err != nil {
						fmt.Printf("[ERR] %s: %v -> sleep 10 sec\n", c.conf.Name, err)
						time.Sleep(10 * time.Second)
						continue
					}
					connAddr := conn.RemoteAddr().String()
					confAddr := c.conf.Remote.String()
					if connAddr != confAddr {
						fmt.Printf("[INF] %s: %s -> ignored \n", c.conf.Name, conn.RemoteAddr().String())
						conn.Close()
						continue
					}
					c.Conn = conn

				case "udp":
					addrSelf := net.UDPAddrFromAddrPort(c.conf.Self)
					addrRemote := net.UDPAddrFromAddrPort(c.conf.Remote)
					conn, err := net.DialUDP("udp", addrSelf, addrRemote)
					if err != nil {
						fmt.Printf("[ERR] %s: %v -> sleep 10 sec\n", c.conf.Name, err)
						time.Sleep(10 * time.Second)
						continue
					}
					c.Conn = conn
				default:
					fmt.Printf("[ERR] %s: Network not supported %s\n", c.conf.Name, c.conf.Network)
					return
				}
				fmt.Printf("[INF] %s: connected\n", c.conf.Name)
			}
			time.Sleep(1 * time.Second)
		}
	}
}
