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

func File(ctx context.Context, logger *log.Logger, h *opsv1.Host, hostOpt option.HostOption, fileOpt option.FileOption) (output string, err error) {
	h.FilledByOption(hostOpt)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return
	}
	return c.File(ctx, fileOpt)
}

func Shell(ctx context.Context, logger *log.Logger, h *opsv1.Host, option option.ShellOption, hostOption option.HostOption) (err error) {
	logger.Info.Println("> Run Shell on ", h.Spec.Address)
	h.FilledByOption(hostOption)
	c, err := NewHostConnBase64(h)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	stdout, err := c.Shell(ctx, option.Sudo, option.Content)
	if err != nil {
		logger.Error.Println(err)
	} else {
		logger.Info.Println(stdout)
	}
	return
}

func GetHosts(logger *log.Logger, clusterOpt option.ClusterOption, hostOpt option.HostOption, inventory string) (hosts []*opsv1.Host) {
	hs, _ := utils.AnalysisHostsParameter(inventory)
	for _, addr := range hs {
		hosts = append(hosts, opsv1.NewHost(clusterOpt.Namespace, strings.ReplaceAll(addr, ".", "-"), addr, hostOpt.Port, hostOpt.Username, hostOpt.Password, hostOpt.PrivateKey, hostOpt.PrivateKeyPath, constants.DefaultSSHTimeoutSeconds, hostOpt.SecretRef))
	}
	return
}
