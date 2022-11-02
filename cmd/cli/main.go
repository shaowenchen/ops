package main

import (
	"flag"

	"github.com/shaowenchen/ops/cmd/cli/root"
	"github.com/spf13/pflag"
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	root.Execute()
}
