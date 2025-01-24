package main

import (
	"context"
	"log"
	"os"

	"github.com/rusl222/gateway/netway"
	"github.com/rusl222/gateway/wintty"
)

func main() {

	print(os.Args[0])

	wt := wintty.Wintty{}
	if err := wt.Read("./wintty.cnf"); err != nil {
		log.Fatal(err)
	}

	master, _ := wt.TTY[0].(wintty.IpDirection)
	slave1, _ := wt.TTY[1].(wintty.IpDirection)
	slave2, _ := wt.TTY[2].(wintty.IpDirection)

	conf := netway.NetwayConfig{
		Master: netway.IpDirection(master),
		Slaves: []netway.IpDirection{
			netway.IpDirection(slave1),
			netway.IpDirection(slave2),
		},
	}

	nw := netway.New(conf)
	ctx := context.Background()
	nw.Run(ctx)
}
