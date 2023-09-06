package constants

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

const LocalHostIP = "127.0.0.1"
const DefaultTimeoutSeconds = 10

const (
	InventoryTypeKubernetes = "kubernetes"
	InventoryTypeHosts      = "hosts"
)

const (
	RemoteStorageTypeImage = "image"
	RemoteStorageTypeS3    = "s3"
	RemoteStorageTypeLocal = "local"
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
	return filepath.Join(GetCurrentUserHomeDir(), ".ops")
}

func GetOpsLogFile() string {
	return filepath.Join(GetOpsLogsDir(), fmt.Sprintf("%d-%d-%d.log", time.Now().Year(), time.Now().Month(), time.Now().Day()))
}

func GetOpsTaskDir() string {
	return filepath.Join(GetOpsDir(), "task")
}

func GetOpsLogsDir() string {
	return filepath.Join(GetOpsDir(), "logs")
}

func GetCurrentUserPrivateKeyPath() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".ssh", "id_rsa")
}
