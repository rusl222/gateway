package main

import (
	"context"
	"net/netip"

	"github.com/rusl222/gateway/netway"
)

func main() {

	master := netway.IpDirection{
		Network: "tcp",
		Role:    "s",
		Self:    netip.MustParseAddrPort("127.0.0.1:4001"),
		Remote:  netip.MustParseAddrPort("127.0.0.1:5001"),
		Comment: "Master",
	}

	slave1 := netway.IpDirection{
		Network: "tcp",
		Role:    "m",
		Self:    netip.MustParseAddrPort("127.0.0.1:4002"),
		Remote:  netip.MustParseAddrPort("127.0.0.1:5002"),
		Comment: "slave1",
	}

	slave2 := netway.IpDirection{
		Network: "udp",
		Self:    netip.MustParseAddrPort("127.0.0.1:4003"),
		Remote:  netip.MustParseAddrPort("127.0.0.1:5003"),
		Comment: "slave2",
	}

	nw := netway.New(netway.NetwayConfig{
		Master: master,
		Slaves: []netway.IpDirection{slave1, slave2},
	})

	ctx := context.Background()
	nw.Run(ctx)
}
