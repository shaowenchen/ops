package host

import (
	"fmt"
	"os"
	"strings"

	"github.com/shaowenchen/opscli/pkg/script"
)

func ActionGetKubeconfig(option KubeconfigOption) (err error) {
	if option.Clear {
		err = os.Remove(GetCurrentUserKubeConfigPath())
		if err != nil {
			return PrintError(err.Error())
		}
	}
	host, err := newHost("", option.Hosts, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
	if err != nil {
		return PrintError(ErrorConnect(err))
	}
	_, err = host.pullContent(GetAdminKubeConfigPath(), GetCurrentUserKubeConfigPath())
	if err != nil {
		PrintError(err.Error())
	}
	return
}

func ActionFile(option FileOption) (err error) {
	if len(option.Hosts) == 0 {
		fmt.Println("remote hosts is empty")
		return
	}
	host, err := newHost("", option.Hosts, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
	if err != nil {
		return PrintError(ErrorConnect(err))
	}
	var size string
	if strings.ToLower(option.Direction) == "download" {
		size, err = host.pull(option.RemoteFile, option.LocalFile)
	} else {
		size, err = host.push(option.LocalFile, option.RemoteFile)
	}
	if err != nil {
		PrintError(err.Error())
	}
	PrintInfo(size)
	return
}

func ActionEtcHosts(option EtcHostsOption) (err error) {
	batchRunHost(option.Hosts, option.Username, option.PrivateKeyPath, script.AddHost(option.IP, option.Domain), script.DeleteHost(option.Domain), option.Clear)
	return nil
}

func ActionInstall(option InstallOption) (err error) {
	if strings.ToLower(option.Name) == "metrics-server" {
		addShell := script.AddlMetricsServer()
		removeShell := script.RemoveMetricsServer()
		batchRunHost(option.Hosts, option.Username, option.PrivateKeyPath, addShell, removeShell, option.Clear)
	}
	return
}

func ActionScript(option ScriptOption) (err error) {
	batchRunHost(option.Hosts, option.Username, option.PrivateKeyPath, script.GetExecutableScript(option.Content), "", false)
	return nil
}

func batchRunHost(hosts, username, privatekeypath, addshell, removeshell string, clear bool) {
	if len(hosts) == 0 {
		hosts = LocalHostIP
	}
	var stdout string
	for _, addr := range RemoveDuplicates(GetSliceFromFileOrString(hosts)) {
		fmt.Printf("host -> %s\n", addr)
		host, err := newHost("", addr, "", 22, username, "", "", privatekeypath, 0)
		if err != nil {
			panic(err)
		}
		if clear {
			stdout, _, err = host.exec(removeshell)
		} else {
			stdout, _, err = host.exec(addshell)
		}
		if len(stdout) != 0 {
			PrintInfo(stdout)
		}
		if err != nil {
			PrintError(ErrorCommon(err))
		}
	}
}
