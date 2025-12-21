package utils

import (
	"bufio"
	"github.com/shaowenchen/ops/pkg/constants"
	"os"
	"strings"
)

func GetInventoryType(inventory string, nodename string) (inventoryType string, availableInventory string) {
	availableInventory = GetAbsoluteFilePath(inventory)
	if nodename != "" {
		inventoryType = constants.InventoryTypeKubernetes
		if availableInventory == "" {
			if IsExistsFile(constants.GetCurrentUserKubeConfigPath()) {
				availableInventory = constants.GetCurrentUserKubeConfigPath()
				return
			} else if IsExistsFile(constants.KubeAdminConfigPath) {
				availableInventory = constants.KubeAdminConfigPath
				return
			}
		} else {
			inventoryType = constants.InventoryTypeKubernetes
			availableInventory = GetAbsoluteFilePath(inventory)
			return
		}
	}
	return constants.InventoryTypeHosts, availableInventory
}

func AnalysisHostsParameter(str string) (result []string, err error) {
	isExist := IsExistsFile(GetAbsoluteFilePath(str))
	if isExist {
		// try kubeconfig
		nodeIPs, err := GetAllNodesByKubeconfig(str)
		if err == nil {
			return nodeIPs, nil
		}
		//try readfile
		readFile, err := os.Open(str)
		if err != nil {
			return result, err
		}
		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			line := strings.TrimSpace(fileScanner.Text())
			line = findIP(line)
			if len(line) > 0 {
				result = append(result, line)
			}
		}
		readFile.Close()
	} else {
		result = SplitStrings(str)
	}
	if len(result) == 0 {
		result = append(result, constants.LocalHostIP)
	}
	return RemoveDuplicates(result), nil
}
