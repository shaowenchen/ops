package main

import (
	"flag"
	"github.com/spf13/pflag"
	"github.com/shaowenchen/ops/cmd/cli/root"
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	root.Execute()
}
