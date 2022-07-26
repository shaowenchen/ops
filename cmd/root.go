package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type Options struct {
	Verbose        bool
	Addons         string
	Name           string
	ClusterCfgPath string
	Kubeconfig     string
	FromCluster    bool
	ClusterCfgFile string
	Kubernetes     string
	Kubesphere     bool
	SkipCheck      bool
	SkipPullImages bool
	KsVersion      string
	Registry       string
	SourcesDir     string
	AddImagesRepo  bool
}

var (
	opt Options
)

var rootCmd = &cobra.Command{
	Use:   "opscli",
	Short: "ops cli",
	Long: `This is a opscli tool for kubernetes cluster.`,
}

func Execute() {
	exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&opt.Verbose, "debug", true, "Print detailed information")
}