package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaowenchen/ops/pkg/option"
	"github.com/spf13/viper"
)

var GlobalConfig = &ConfigOptions{}

type ConfigOptions struct {
	Server  ServerOptions        `mapstructure:"server"`
	Copilot option.CopilotOption `mapstructure:"copilot"`
	Event   EventOption          `mapstructure:"event"`
}

type ServerOptions struct {
	RunMode string `mapstructure:"runmode"`
	Token   string `mapstructure:"token"`
}

type EventOption struct {
	Endpoint string `mapstructure:"endpoint"`
	Cluster  string `mapstructure:"cluster"`
}

func LoadConfig(configPath string) {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if configPath == "" {
		viper.SetConfigName("default")
		viper.SetConfigType("toml")
	} else {
		viper.SetConfigFile(configPath)
	}
	viper.AddConfigPath(filepath.Join(path, "."))
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Printf("fatal error config file: %s \n", err)
	}
	err = viper.Unmarshal(GlobalConfig)
	if err != nil {
		fmt.Printf("unmarshal config file: %s \n", err)
	}
}
