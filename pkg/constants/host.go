package constants

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
)

const LocalHostIP = "127.0.0.1"
const DefaultSSHTimeoutSeconds = 30
const DefaultShellTimeoutSeconds = 30
const DefaultShellTimeoutDuration = DefaultShellTimeoutSeconds * time.Second

const (
	InventoryTypeKubernetes = "kubernetes"
	InventoryTypeHosts      = "hosts"
)

const (
	RemoteStorageTypeS3     = "s3"
	RemoteStorageTypeImage  = "image"
	RemoteStorageTypeServer = "server"
)

func GetOsInfo() string {
	cmd := exec.Command("uname", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return out.String()
}

func GetCurrentUserHomeDir() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return os.TempDir()
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
	return filepath.Join(GetOpsDir(), "tasks")
}

func GetCloudTaskDir() string {
	return "https://raw.githubusercontent.com/shaowenchen/ops/main/tasks"
}

func GetOpsLogsDir() string {
	return filepath.Join(GetOpsDir(), "logs")
}

func GetCurrentUserPrivateKeyPath() string {
	return filepath.Join(GetCurrentUserHomeDir(), ".ssh", "id_rsa")
}
