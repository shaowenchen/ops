package host

import (
	"errors"
	"fmt"
	"github.com/shaowenchen/ops/pkg/log"
	"github.com/shaowenchen/ops/pkg/utils"
)

func ActionBatchFile(logger *log.Logger, option FileOption) (err error) {
	hosts, _ := utils.AnalysisHostsParameter(option.Hosts)
	if utils.IsDownloadDirection(option.Direction) && len(hosts) != 1 {
		errMsg := "need only one host while downloading"
		logger.Error.Println(errMsg)
		return errors.New(errMsg)
	}
	for _, addr := range hosts {
		logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", addr)))
		host, err := NewHost(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		err = host.File(logger, option.Sudo, option.Direction, option.LocalFile, option.RemoteFile)
	}
	return
}

func ActionBatchScript(logger *log.Logger, option ScriptOption) (err error) {
	hosts, err := utils.AnalysisHostsParameter(option.Hosts)
	for _, addr := range hosts {
		logger.Info.Print(utils.PrintMiddleFilled(fmt.Sprintf("[%s]", addr)))
		host, err := NewHost(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
		if err != nil {
			logger.Error.Println(err)
			continue
		}
		host.Script(logger, option.Sudo, option.Content)
	}
	return
}
