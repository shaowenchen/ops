package constants

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

func GetCurrentUserHomeDir() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDirectory
}

func GetCurrentUser() string {
	currentUser, err := user.Current()
	if err != nil {
		panic(err)
	}
	return currentUser.Username
}

func GetOpsDir() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".opscli")
}

func GetOpscliLogFile() string {
	return filepath.Join(GetOpscliLogsDir(), fmt.Sprintf("%d-%d-%d.log", time.Now().Year(), time.Now().Month(), time.Now().Day()))
}

func GetOpscliPipelineDir() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".opscli", "pipeline")
}

func GetOpscliLogsDir() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".opscli", "logs")
}

func GetCurrentUserPrivateKeyPath() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".ssh", "id_rsa")
}
