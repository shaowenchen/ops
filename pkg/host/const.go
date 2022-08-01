package host

import (
	"os"
	"path/filepath"
)

func GetCurrentUserHomeDir() string {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return homeDirectory
}

func GetCurrentUserPrivateKeyPath() string {
	homeDirectory := GetCurrentUserHomeDir()
	return filepath.Join(homeDirectory, ".ssh", "id_rsa")
}

func GetCurrentUserKubeConfigPath() string {
	homeDirectory := GetCurrentUserHomeDir()
	return filepath.Join(homeDirectory, ".kube", "config")
}

func GetAdminKubeConfigPath() string {
	return "/etc/kubernetes/admin.conf"
}
