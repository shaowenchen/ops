package host

import (
	"github.com/shaowenchen/opscli/pkg/script"
	"os"
)

func ActionGetKubeconfig(option KubeconfigOption) (err error) {
	if option.Clear {
		err = os.Remove(GetCurrentUserKubeConfigPath())
		if err != nil {
			return PrintError(err.Error())
		}
	}
	host, err := newHost("", option.Input, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
	if err != nil {
		return PrintError(ErrorConnect(err))
	}
	err = host.Fetch(GetAdminKubeConfigPath(), GetCurrentUserKubeConfigPath())
	if err != nil {
		PrintError(err.Error())
	}
	return
}

func ActionEtcHosts(option EtcHostsOption) (err error) {
	for _, addr := range SplitStr(option.Input) {
		host, err := newHost("", addr, "", 22, option.Username, "", "", option.PrivateKeyPath, 0)
		if err != nil {
			return PrintError(ErrorConnect(err))
		}
		if option.Clear {
			_, _, err = host.exec(script.DeleteHost(option.Domain))
		} else {
			_, _, err = host.exec(script.AddHost(option.IP, option.Domain))
		}
		if err != nil {
			PrintError(ErrorEtcHosts(err))
		}
	}
	return nil
}
