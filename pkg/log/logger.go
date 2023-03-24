package log

import (
	"bytes"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"io"
	"log"
	"os"
)

const (
	LevelFatal = iota
	LevelError
	LevelWarning
	LevelInfo
	LevelDebug
	LevelTrace
)

type Logger struct {
	PrintLog   bool
	FileLog    bool
	BufferLog  bool
	Trace      *log.Logger
	Debug      *log.Logger
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	Fatal      *log.Logger
	BufferData *bytes.Buffer
	Level      int
}

func NewCliLogger(printLog bool, fileLog bool, level int) (*Logger, error) {
	return NewLogger(true, true, false, level)
}

func NewServerLogger(printLog bool, bufferLog bool, level int) (*Logger, error) {
	return NewLogger(true, false, true, level)
}

func NewLogger(printLog bool, fileLog bool, bufferLog bool, level int) (*Logger, error) {
	l := &Logger{
		PrintLog:  printLog,
		FileLog:   fileLog,
		BufferLog: bufferLog,
		Level:     level,
	}
	err := l.init(level)
	return l, err
}

func (l *Logger) init(level int) (err error) {
	multiWriter := io.MultiWriter()
	if l.PrintLog {
		multiWriter = io.MultiWriter(multiWriter, os.Stdout)
	}
	if l.FileLog {
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
	if l.BufferLog {
		l.BufferData = new(bytes.Buffer)
		multiWriter = io.MultiWriter(multiWriter, l.BufferData)
	}

	if LevelFatal <= level {
		l.Fatal = log.New(multiWriter, "[FATAL]", 0)
	} else {
		l.Fatal = log.New(io.Discard, "[FATAL]", 0)
	}
	if LevelError <= level {
		l.Error = log.New(multiWriter, "[ERROR]", 0)
	} else {
		l.Error = log.New(io.Discard, "[ERROR]", 0)
	}
	if LevelWarning <= level {
		l.Warning = log.New(multiWriter, "[WARNING]", 0)
	} else {
		l.Warning = log.New(io.Discard, "[WARNING]", 0)
	}
	if LevelInfo <= level {
		l.Info = log.New(multiWriter, "[INFO]", 0)
	} else {
		l.Info = log.New(io.Discard, "[INFO]", 0)
	}
	if LevelDebug <= level {
		l.Debug = log.New(multiWriter, "[DEBUG]", 0)
	} else {
		l.Debug = log.New(io.Discard, "[DEBUG]", 0)
	}
	if LevelTrace <= level {
		l.Trace = log.New(multiWriter, "[TRACE]", 0)
	} else {
		l.Trace = log.New(io.Discard, "[TRACE]", 0)
	}
	return
}

func (logger *Logger) GetBuffer() (log string) {
	log = string(logger.BufferData.Bytes())
	return
}
