package main

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"time"
)

func init() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
}

func main() {
	server, err := NewOpsServer()
	if err != nil {
		fmt.Println(err)
		return
	}
	server.Run()
	return
}

type OpsServer struct {
}

func NewOpsServer() (server *OpsServer, err error) {
	server = &OpsServer{}
	return server, nil
}

func (server *OpsServer) Run() (err error) {

	for range time.Tick(time.Second * time.Duration(3)) {
		fmt.Println("OK")
	}
	return
}
