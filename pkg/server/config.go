package server

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

var GlobalConfig = &ConfigOptions{}

type ConfigOptions struct {
	Server ServerOptions
}

type ServerOptions struct {
	RunMode string
	Token   string
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
