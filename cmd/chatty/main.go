package main

import (
	"flag"
	"fmt"

	"github.com/blackchip-org/chatty/irc"
)

var s = irc.Server{}

func init() {
	flag.StringVar(&s.Addr, "address", irc.Addr, "address to listen on")
	flag.StringVar(&s.CertFile, "cert", "chatty.cert", "certificate file for tls")
	flag.StringVar(&s.DataFile, "data", "chatty.data", "file that holds persistent data")
	flag.BoolVar(&s.Debug, "debug", false, "enable debug")
	flag.BoolVar(&s.Insecure, "insecure", false, "use plaintext instead of tls")
	flag.StringVar(&s.KeyFile, "key", "chatty.key", "key file for tls")
	flag.StringVar(&s.Name, "name", irc.ServerName, "override the name of the server")
}

func main() {
	flag.Parse()
	err := s.ListenAndServe()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
