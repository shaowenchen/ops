package action

import (
	"errors"
	"fmt"
	"github.com/shaowenchen/ops/pkg/host"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
)

func HostFile(logger *log.Logger, option host.FileOption) (err error) {
	hosts, _ := utils.AnalysisHostsParameter(option.Hosts)
	if utils.IsDownloadDirection(option.Direction) && len(hosts) != 1 {
		errMsg := "need only one host while downloading"
		logger.Error.Println(errMsg)
		return errors.New(errMsg)
	}
	for _, addr := range hosts {
		logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", addr)))
		c, err := host.NewHostConnection(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		err = c.File(option.Sudo, option.Direction, option.LocalFile, option.RemoteFile)
	}
	return
}

func HostScript(logger *log.Logger, option host.ScriptOption) (err error) {
	hosts, err := utils.AnalysisHostsParameter(option.Hosts)
	for _, addr := range hosts {
		logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", addr)))
		c, err := host.NewHostConnection(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		stdout, _, _ := c.Script(option.Sudo, option.Content)
		logger.Info.Println(stdout)
	}
	return
}
