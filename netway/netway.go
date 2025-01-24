package netway

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	"github.com/rusl222/gateway/concontrol"
)

type NetwayConfig struct {
	Master IpDirection
	Slaves []IpDirection
}

type IpDirection struct {
	Network string
	Role    string //m-master(clent), s-slave(server only conf IP), p-passive(server any IP)
	Self    netip.AddrPort
	Remote  netip.AddrPort
	Comment string
}

type Netway struct {
	conf       NetwayConfig
	connMaster *concontrol.ConController
	connSlaves []*concontrol.ConController
}

func New(conf NetwayConfig) *Netway {
	return &Netway{conf: conf}
}

func (n *Netway) Run(ctx context.Context) error {

	// all slaves
	for _, slave := range n.conf.Slaves {
		slave, err := concontrol.New(
			concontrol.ConControllerConfig{
				Network: slave.Network,
				Role:    slave.Role,
				Self:    slave.Self,
				Remote:  slave.Remote,
				Name:    slave.Comment,
			})
		if err != nil {
			return err
		}
		go slave.Run(ctx)
		n.connSlaves = append(n.connSlaves, slave)
	}

	// new Master
	master, err := concontrol.New(concontrol.ConControllerConfig{
		Network: n.conf.Master.Network,
		Role:    n.conf.Master.Role,
		Self:    n.conf.Master.Self,
		Remote:  n.conf.Master.Remote,
		Name:    n.conf.Master.Comment,
	})
	if err != nil {
		return err
	}
	go master.Run(ctx)
	n.connMaster = master

	// transport
	for _, conn := range n.connSlaves {
		go n.transport(ctx, conn, []*concontrol.ConController{n.connMaster})
	}
	go n.transport(ctx, n.connMaster, n.connSlaves)

	// wait for the context to be done
	<-ctx.Done()
	return nil
}

// transport - transmitting packets from the client to the servers
// and in the revers direction.
func (s *Netway) transport(ctx context.Context, src *concontrol.ConController, dst []*concontrol.ConController) {
	var n int
	var err error
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if src.Conn != nil {
				n, err = src.Conn.Read(buf)
				if err != nil {

					fmt.Printf("[ERR] %s: %v\n", src.Name(), err)
					src.Conn.Close()
					src.Reconnect()
					continue
				}
				for _, c := range dst {
					if c.Conn != nil {
						_, err := c.Conn.Write(buf[:n])
						if err != nil {
							fmt.Printf("[ERR] %s: %v\n", c.Name(), err)
							c.Conn.Close()
							c.Reconnect()
						}
					}
				}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}
