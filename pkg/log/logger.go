package log

import (
	"io"
	"log"
	"os"

	"github.com/shaowenchen/ops/pkg/constants"
)

type Logger struct {
	Print   bool
	LogFile string
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewDefaultLogger(print bool, file bool) (*Logger, error) {
	return NewLogger(constants.GetOpscliLogFile(), true, true)
}

func NewLogger(logFile string, print bool, file bool) (*Logger, error) {
	logger := &Logger{
		LogFile: logFile,
		Print:   print,
	}
	err := logger.init("")
	return logger, err
}

func (logger *Logger) init(prefix string) error {
	file, err := os.OpenFile(logger.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	multiWriter := io.MultiWriter(file)
	if logger.Print {
		multiWriter = io.MultiWriter(file, os.Stdout)
	}
	logger.Info = log.New(multiWriter, prefix, 0)
	logger.Warning = log.New(multiWriter, "", 0)
	logger.Error = log.New(multiWriter, "", 0)
	return err
}
