package main

import (
	"flag"
	"os"

	"github.com/blackchip-org/chatty/irc"
)

var debug bool
var address string
var name string

var hostname = "localhost"

func init() {
	n, err := os.Hostname()
	if err == nil {
		hostname = n
	}
}

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.StringVar(&address, "address", ":6667", "address to listen on")
	flag.StringVar(&name, "name", hostname, "override the name of the server")
}

func main() {
	flag.Parse()
	s := &irc.Server{
		Address: address,
		Debug:   debug,
		Name:    name,
	}
	s.Run()
}
