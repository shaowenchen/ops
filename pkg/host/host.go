package host

import (
	"context"
	opsv1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/option"
	"github.com/shaowenchen/ops/pkg/utils"
	"strings"
)

func File(ctx context.Context, logger *log.Logger, h *opsv1.Host, fileOpt option.FileOption, hostOption option.HostOption) (err error) {
	FillHostByOption(h, &hostOption)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	return c.File(ctx, fileOpt.Sudo, fileOpt.Direction, fileOpt.LocalFile, fileOpt.RemoteFile)
}

func Shell(ctx context.Context, logger *log.Logger, h *opsv1.Host, option option.ShellOption, hostOption option.HostOption) (err error) {
	logger.Info.Println("> Run Shell on ", h.Spec.Address)
	FillHostByOption(h, &hostOption)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	stdout, err := c.Shell(ctx, option.Sudo, option.Content)
	logger.Info.Println(stdout)
	return
}

func GetHosts(logger *log.Logger, clusterOpt option.ClusterOption, hostOpt option.HostOption, inventory string) (hosts []*opsv1.Host) {
	hs, _ := utils.AnalysisHostsParameter(inventory)
	for _, addr := range hs {
		hosts = append(hosts, opsv1.NewHost(clusterOpt.Namespace, strings.ReplaceAll(addr, ".", "-"), addr, hostOpt.Port, hostOpt.Username, hostOpt.Password, hostOpt.PrivateKey, hostOpt.PrivateKeyPath, constants.DefaultSSHTimeoutSeconds, hostOpt.SecretRef))
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
