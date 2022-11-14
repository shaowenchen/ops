package host

import (
	"fmt"

	v1 "github.com/shaowenchen/ops/api/v1"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
)

func File(logger *log.Logger, h *v1.Host, option FileOption) (err error) {
	logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", h.Spec.Address)))
	c, err := NewHostConnection(h.Spec.Address, option.Port, option.Username, option.Password, option.PrivateKey, option.PrivateKeyPath)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	return c.File(option.Sudo, option.Direction, option.LocalFile, option.RemoteFile)
}

func Script(logger *log.Logger, h *v1.Host, option ScriptOption) (err error) {
	logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", h.Spec.Address)))
	c, err := NewHostConnection(h.Spec.Address, option.Port, option.Username, option.Password, option.PrivateKey, option.PrivateKeyPath)
	if err != nil {
		logger.Error.Println(err)
		return err
	}
	stdout, _, err := c.Script(option.Sudo, option.Content)
	logger.Info.Println(stdout)
	return
}

func GetHosts(logger *log.Logger, option HostOption) (hosts []*v1.Host) {
	hs, _ := utils.AnalysisHostsParameter(option.Hosts)
	for _, addr := range hs {
		hosts = append(hosts, v1.NewHost("", "", addr, option.Port, option.Username, option.Password, "", option.PrivateKeyPath))
	}
	return
}
