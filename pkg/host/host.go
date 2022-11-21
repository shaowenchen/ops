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
	logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", h.Spec.Address)))
	c, err := NewHostConnectionBase64(h.Spec.Address, hostOption.Port, hostOption.Username, hostOption.Password, hostOption.PrivateKey, hostOption.PrivateKeyPath)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	return c.File(option.Sudo, option.Direction, option.LocalFile, option.RemoteFile)
}

func Script(logger *log.Logger, h *opsv1.Host, option option.ScriptOption, hostOption option.HostOption) (err error) {
	logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", h.Spec.Address)))
	c, err := NewHostConnectionBase64(h.Spec.Address, hostOption.Port, hostOption.Username, hostOption.Password, hostOption.PrivateKey, hostOption.PrivateKeyPath)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	stdout, _, err := c.Script(option.Sudo, option.Script)
	logger.Info.Println(stdout)
	return
}

func GetHosts(logger *log.Logger, option option.HostOption, inventory string) (hosts []*opsv1.Host) {
	hs, _ := utils.AnalysisHostsParameter(inventory)
	for _, addr := range hs {
		hosts = append(hosts, opsv1.NewHost("default", strings.ReplaceAll(addr, ".", "-"), addr, option.Port, option.Username, option.Password, option.PrivateKey, option.PrivateKeyPath, constants.DefaultTimeoutSeconds))
	}
	return
}
