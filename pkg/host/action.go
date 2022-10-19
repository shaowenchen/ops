package host

import (
	"fmt"
	"os"
	"strings"

	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionGetKubeconfig(option KubeconfigOption) (err error) {
	if option.Clear {
		err = os.Remove(GetCurrentUserKubeConfigPath())
		if err != nil {
			return utils.PrintError(err)
		}
	}
	host, err := newHost("", option.Hosts, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
	if err != nil {
		return utils.PrintError(err.Error())
	}
	err = host.pullContent(GetAdminKubeConfigPath(), GetCurrentUserKubeConfigPath())
	if err != nil {
		utils.PrintError(err)
	}
	return
}

func ActionFile(option FileOption) (err error) {
	hosts := RemoveDuplicates(GetSliceFromFileOrString(option.Hosts))
	if strings.ToLower(option.Direction) == "download" {
		hosts := RemoveDuplicates(GetSliceFromFileOrString(option.Hosts))
		if len(hosts) != 1 {
			fmt.Println("need only one target host")
			return
		}
		host, err := newHost("", hosts[0], "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
		if err != nil {
			return utils.PrintError(err)
		}
		md5, err := host.fileMd5(option.RemoteFile)
		if err != nil {
			return utils.PrintError(err)
		}
		err = host.pull(option.RemoteFile, option.LocalFile, md5)
		if err != nil {
			return utils.PrintError(err)
		}
		utils.PrintInfo("Md5: ", md5)
	} else {
		md5, err := FileMD5(option.LocalFile)
		if err != nil {
			return utils.PrintError(err)
		}
		for _, addr := range hosts {
			host, err := newHost("", addr, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
			if err != nil {
				return utils.PrintError(err)
			}
			err = host.push(option.LocalFile, option.RemoteFile, md5)
			if err != nil {
				utils.PrintError(err)
			}
			utils.PrintInfo("Md5: ", md5)
		}
	}
	return
}

func ActionEtcHosts(option EtcHostsOption) (err error) {
	batchRunHost(option.Hosts, option.Username, option.PrivateKeyPath, utils.AddHost(option.IP, option.Domain), utils.DeleteHost(option.Domain), option.Clear)
	return nil
}

func ActionScript(option ScriptOption) (err error) {
	batchRunHost(option.Hosts, option.Username, option.PrivateKeyPath, utils.GetExecutableScript(option.Content), "", false)
	return nil
}

func batchRunHost(hosts, username, privatekeypath, addshell, removeshell string, clear bool) {
	if len(hosts) == 0 {
		hosts = LocalHostIP
	}
	var stdout string
	for _, addr := range RemoveDuplicates(GetSliceFromFileOrString(hosts)) {
		host, err := newHost("", addr, "", 22, username, "", "", privatekeypath, 0)
		if err != nil {
			utils.PrintError(err)
			continue
		}
		if clear {
			stdout, _, err = host.exec(removeshell)
		} else {
			stdout, _, err = host.exec(addshell)
		}
		if len(stdout) != 0 {
			utils.PrintInfo(fmt.Sprintf("[%s] %s", addr, stdout))
		}
		if err != nil {
			utils.PrintError(err)
		}
	}
}
