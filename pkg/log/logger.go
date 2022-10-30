package log

import (
	"io"
	"log"
	"os"

	"github.com/shaowenchen/opscli/pkg/constants"
)

type Logger struct {
	Print   bool
	LogFile string
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
}

func NewDefaultLogger(print bool) (*Logger, error) {
	return NewLogger(constants.GetOpscliLogFile(), true)
}

func NewLogger(logFile string, print bool) (*Logger, error) {
	logger := &Logger{
		LogFile: logFile,
		Print:   print,
	}
	err := logger.init()
	return logger, err
}

func (logger *Logger) init() error {
	file, err := os.OpenFile(logger.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	multiWriter := io.MultiWriter(file)
	if logger.Print {
		multiWriter = io.MultiWriter(file, os.Stdout)
	}
	logger.Info = log.New(multiWriter, "[INFO]", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Warning = log.New(multiWriter, "[WARNING]", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Error = log.New(multiWriter, "[ERROR]", log.Ldate|log.Ltime|log.Lshortfile)
	return err
}
