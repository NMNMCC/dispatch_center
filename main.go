package main

import (
	"time"

	"github.com/alecthomas/kong"
	"nmnm.cc/dispatch_center/internal/server"
)

var cli struct {
	Listen               string        `help:"Address to listen on" short:"l"`
	DatabasePath         string        `help:"Path to the database" short:"d"`
	NodeKeepaliveTimeout time.Duration `help:"Node keepalive timeout" default:"30s"`
}

func main() {
	var _ = kong.Parse(&cli)

	s, err := server.New(cli.DatabasePath, cli.NodeKeepaliveTimeout)
	if err != nil {
		panic(err)
	}

	s.Start(cli.Listen)
}
