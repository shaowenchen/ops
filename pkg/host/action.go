package host

import (
	"fmt"
	"os"
	"strings"

	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionGetKubeconfig(option KubeconfigOption) (err error) {
	if option.Clear {
		err = os.Remove(utils.GetCurrentUserKubeConfigPath())
		if err != nil {
			return
		}
	}
	host, err := newHost(option.Hosts, option.Port, option.Username, option.Password, option.PrivateKeyPath)
	if err != nil {
		return
	}
	err = host.pullContent(utils.GetAdminKubeConfigPath(), utils.GetCurrentUserKubeConfigPath())
	if err != nil {
		return
	}
	return
}

func ActionFile(option FileOption) (err error) {
	hosts := utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts))
	if strings.ToLower(option.Direction) == "download" {
		hosts := utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts))
		if len(hosts) != 1 {
			return utils.LogError("need only one target host")
		}
		host, err := newHost(hosts[0], option.Port, option.Username, option.Password, option.PrivateKeyPath)

		if err != nil {
			return utils.LogError(err)
		}
		md5, err := host.fileMd5(option.RemoteFile)
		if err != nil {
			return utils.LogError(err)
		}
		err = host.pull(option.RemoteFile, option.LocalFile, md5)
		if err != nil {
			return utils.LogError(err)
		}
		utils.LogInfo("Md5: ", md5)
	} else {
		md5, err := utils.FileMD5(option.LocalFile)
		if err != nil {
			return utils.LogError(err)
		}
		for _, addr := range hosts {
			host, err := newHost(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
			if err != nil {
				return utils.LogError(err)
			}
			err = host.push(option.LocalFile, option.RemoteFile, md5)
			if err != nil {
				utils.LogError(err)
			}
			utils.LogInfo("Md5: ", md5)
		}
	}
	return
}

func ActionBatchScript(option ScriptOption) (err error) {
	if len(option.Hosts) == 0 {
		option.Hosts = LocalHostIP
	}
	for _, addr := range utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts)) {
		_, _, err = runHost(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath, option.Content)
	}
	return
}

func ActionScript(option ScriptOption) (stdout string, exit int, err error) {
	if len(option.Hosts) == 0 {
		option.Hosts = LocalHostIP
	}
	return runHost(option.Hosts, option.Port, option.Username, option.Password, option.PrivateKeyPath, option.Content)
}

func runHost(addr string, port int, username, password, privatekeypath, shell string) (stdout string, exit int, err error) {
	host, err := newHost(addr, port, username, password, privatekeypath)
	if err != nil {
		return "", 1, utils.LogError(err)
	}
	stdout, exit, err = host.exec(shell)
	if len(stdout) != 0 {
		utils.LogInfo(fmt.Sprintf("[%s] %s", addr, stdout))
	}
	if exit != 0 {
		return "", 1, utils.LogError(err)
	}
	return
}
