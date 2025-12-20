package config

import (
	"fmt"
	"os"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// AllowedConfigKeys defines the allowed configuration keys
var AllowedConfigKeys = map[string]bool{
	"runtimeimage": true,
	"proxy":        true,
}

// AllowedConfigKeysOrder defines the order of configuration keys for display
var AllowedConfigKeysOrder = []string{
	"proxy",
	"runtimeimage",
}

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "configure opscli settings",
	Long:  `Configure opscli settings. Settings are stored in ~/.ops/opscli/config`,
}

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "set a configuration value",
	Long:  `Set a configuration value. Allowed keys: runtimeimage, proxy. Example: opscli config set proxy https://proxy.example.com`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		if !isAllowedKey(key) {
			fmt.Fprintf(os.Stderr, "Error: Invalid configuration key '%s'. Allowed keys are: runtimeimage, proxy\n", key)
			os.Exit(1)
		}
		if err := setConfig(key, value); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration '%s' set to '%s'\n", key, value)
	},
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "get a configuration value",
	Long:  `Get a configuration value. Example: opscli config get proxy`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value, err := getConfig(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting config: %v\n", err)
			os.Exit(1)
		}
		if value == "" {
			fmt.Fprintf(os.Stderr, "Configuration key '%s' not found\n", key)
			os.Exit(1)
		}
		fmt.Println(value)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list all configuration values",
	Long:  `List all configuration values`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
		// List all allowed keys with their values
		for _, key := range AllowedConfigKeysOrder {
			value := config[key]
			if value == "" {
				fmt.Printf("%s = (not set)\n", key)
			} else {
				fmt.Printf("%s = %s\n", key, value)
			}
		}
	},
}

var unsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "unset a configuration value",
	Long:  `Unset a configuration value. Allowed keys: runtimeimage, proxy. Example: opscli config unset proxy`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if !isAllowedKey(key) {
			fmt.Fprintf(os.Stderr, "Error: Invalid configuration key '%s'. Allowed keys are: runtimeimage, proxy\n", key)
			os.Exit(1)
		}
		if err := unsetConfig(key); err != nil {
			fmt.Fprintf(os.Stderr, "Error unsetting config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Configuration '%s' unset\n", key)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
	ConfigCmd.AddCommand(getCmd)
	ConfigCmd.AddCommand(listCmd)
	ConfigCmd.AddCommand(unsetCmd)
}

func loadConfig() (map[string]string, error) {
	configPath := constants.GetOpsCliConfigPath()
	config := make(map[string]string)

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		return config, nil
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func saveConfig(config map[string]string) error {
	configPath := constants.GetOpsCliConfigPath()
	configDir := constants.GetOpsCliConfigDir()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func setConfig(key, value string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	config[key] = value
	return saveConfig(config)
}

func getConfig(key string) (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", err
	}

	return config[key], nil
}

func unsetConfig(key string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	if _, exists := config[key]; !exists {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	delete(config, key)
	return saveConfig(config)
}

func isAllowedKey(key string) bool {
	return AllowedConfigKeys[key]
}

// GetConfigValue returns the configuration value for the given key.
// Returns empty string if the key is not set or not allowed.
func GetConfigValue(key string) string {
	if !isAllowedKey(key) {
		return ""
	}
	value, err := getConfig(key)
	if err != nil {
		return ""
	}
	return value
}

// GetValueWithPriority returns the value following priority:
// 1. CLI argument (if provided and not empty)
// 2. Environment variable
// 3. Config file
// 4. Default value
func GetValueWithPriority(cliValue, envKey, configKey, defaultValue string) string {
	// Priority 1: CLI argument
	if cliValue != "" && cliValue != defaultValue {
		return cliValue
	}

	// Priority 2: Environment variable
	if envKey != "" {
		if envValue := os.Getenv(envKey); envValue != "" {
			return envValue
		}
	}

	// Priority 3: Config file
	if configKey != "" {
		if configValue := GetConfigValue(configKey); configValue != "" {
			return configValue
		}
	}

	// Priority 4: Default value
	return defaultValue
}
