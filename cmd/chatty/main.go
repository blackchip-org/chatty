package main

import (
	"flag"

	"github.com/blackchip-org/chatty/irc"
)

var debug bool
var addr string
var name string

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.StringVar(&addr, "address", irc.Addr, "address to listen on")
	flag.StringVar(&name, "name", irc.ServerName, "override the name of the server")
}

func main() {
	flag.Parse()
	s := &irc.Server{
		Addr:  addr,
		Debug: debug,
		Name:  name,
	}
	s.Run()
}
