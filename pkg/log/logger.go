package log

import (
	"bytes"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"io"
	"log"
	"os"
)

type Logger struct {
	PrintLog   bool
	FileLog    bool
	BufferLog  bool
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	BufferData *bytes.Buffer
}

func NewCliLogger(printLog bool, fileLog bool) (*Logger, error) {
	return NewLogger(true, true, false)
}

func NewServerLogger(printLog bool, bufferLog bool) (*Logger, error) {
	return NewLogger(true, false, true)
}

func NewLogger(printLog bool, FileLog bool, BufferLog bool) (*Logger, error) {
	logger := &Logger{
		PrintLog:  printLog,
		FileLog:   FileLog,
		BufferLog: BufferLog,
	}
	err := logger.init("")
	return logger, err
}

func (logger *Logger) init(prefix string) (err error) {
	multiWriter := io.MultiWriter()
	if logger.PrintLog {
		multiWriter = io.MultiWriter(multiWriter, os.Stdout)
	}
	if logger.FileLog {
		err := utils.CreateDir(constants.GetOpsLogsDir())
		if err != nil {
			return err
		}
		file, err := os.OpenFile(constants.GetOpsLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		multiWriter = io.MultiWriter(multiWriter, file)
	}
	if logger.BufferLog {
		logger.BufferData = new(bytes.Buffer)
		multiWriter = io.MultiWriter(multiWriter, logger.BufferData)
	}
	logger.Info = log.New(multiWriter, prefix, 0)
	logger.Warning = log.New(multiWriter, "", 0)
	logger.Error = log.New(multiWriter, "", 0)
	return
}

func (logger *Logger) GetBuffer() (log string) {
	log = string(logger.BufferData.Bytes())
	return
}
