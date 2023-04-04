package host

import (
	"fmt"
	"strings"

	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
)

func File(logger *log.Logger, h *opsv1.Host, option option.FileOption, hostOption option.HostOption) (err error) {
	logger.Info.Print(utils.FilledInMiddle(fmt.Sprintf("[%s]", h.Spec.Address)))
	FillHostByOption(h, &hostOption)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	return c.File(option.Sudo, option.Direction, option.LocalFile, option.RemoteFile)
}

func Shell(logger *log.Logger, h *opsv1.Host, option option.ShellOption, hostOption option.HostOption) (err error) {
	logger.Info.Print(utils.FilledInMiddle(fmt.Sprintf("[%s]", h.Spec.Address)))
	FillHostByOption(h, &hostOption)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	stdout, err := c.Shell(option.Sudo, option.Content)
	logger.Info.Println(stdout)
	return
}

func GetHosts(logger *log.Logger, option option.HostOption, inventory string) (hosts []*opsv1.Host) {
	hs, _ := utils.AnalysisHostsParameter(inventory)
	for _, addr := range hs {
		hosts = append(hosts, opsv1.NewHost(constants.DefaultNamespace, strings.ReplaceAll(addr, ".", "-"), addr, option.Port, option.Username, option.Password, option.PrivateKey, option.PrivateKeyPath, constants.DefaultTimeoutSeconds))
	}
	return
}

func FillHostByOption(h *opsv1.Host, option *option.HostOption) *opsv1.Host {
	if option.Username != "" && h.GetSpec().Username == "" {
		h.Spec.Username = option.Username
	}
	if option.Password != "" && h.GetSpec().Password == "" {
		h.Spec.Password = option.Password
	}
	if option.Port != 0 && h.GetSpec().Port == 0 {
		h.Spec.Port = option.Port
	}
	if option.PrivateKey != "" && h.GetSpec().PrivateKey == "" {
		h.Spec.PrivateKey = option.PrivateKey
	}
	if option.PrivateKeyPath != "" && h.GetSpec().PrivateKeyPath == "" {
		h.Spec.PrivateKeyPath = option.PrivateKeyPath
	}
	return h
}
