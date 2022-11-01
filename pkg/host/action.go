package host

import (
	"github.com/shaowenchen/opscli/pkg/log"
	"github.com/shaowenchen/opscli/pkg/utils"
)

func ActionFile(logger *log.Logger, option FileOption) (err error) {
	hosts := utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts))
	option.LocalFile = utils.GetAbsoluteFilePath(option.LocalFile)
	isExist, _ := utils.IsExistsFile(option.LocalFile)
	if !isExist {
		hosts := utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts))
		if len(hosts) != 1 {
			logger.Error.Println("need only one target host")
			return
		}
		host, err := newHost(hosts[0], option.Port, option.Username, option.Password, option.PrivateKeyPath)

		if err != nil {
			logger.Error.Println(err)
			return err
		}
		md5, err := host.fileMd5(option.RemoteFile)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		err = host.pull(option.RemoteFile, option.LocalFile, md5)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		logger.Info.Println("Md5: ", md5)
	} else {
		md5, err := utils.FileMD5(option.LocalFile)
		if err != nil {
			logger.Error.Println(err)
			return err
		}
		for _, addr := range hosts {
			host, err := newHost(addr, option.Port, option.Username, option.Password, option.PrivateKeyPath)
			if err != nil {
				logger.Error.Println(err)
				return err
			}
			err = host.push(option.LocalFile, option.RemoteFile, md5)
			if err != nil {
				logger.Error.Println(err)
			}
			logger.Info.Println("Md5: ", md5)
		}
	}
	return
}

func ActionBatchScript(logger *log.Logger, option ScriptOption) (err error) {
	if len(option.Hosts) == 0 {
		option.Hosts = LocalHostIP
	}
	for _, addr := range utils.RemoveDuplicates(utils.GetSliceFromFileOrString(option.Hosts)) {
		scriptOption := option
		scriptOption.Hosts = addr
		_, _, err = ActionScript(logger, scriptOption)
	}
	return
}

func ActionScript(logger *log.Logger, option ScriptOption) (stdout string, exit int, err error) {
	if len(option.Hosts) == 0 {
		option.Hosts = LocalHostIP
	}
	host, err := newHost(option.Hosts, option.Port, option.Username, option.Password, option.PrivateKeyPath)
	if err != nil {
		logger.Error.Println(err)
		return "", 1, err
	}
	stdout, exit, err = host.exec(option.Content)
	if len(stdout) != 0 {
		logger.Info.Println(stdout)
	}
	if exit != 0 {
		logger.Error.Println(err)
		return "", 1, err
	}
	return
}
