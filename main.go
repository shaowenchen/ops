package main

import (
	"flag"
	"github.com/shaowenchen/opscli/cmd"
	"github.com/spf13/pflag"
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	cmd.Execute()
}
