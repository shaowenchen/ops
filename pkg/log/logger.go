package log

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
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
	FileLog    bool
	StdLog     bool
	Plain      bool
	Trace      *log.Logger
	Debug      *log.Logger
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	Fatal      *log.Logger
	BufferData *bytes.Buffer
	Level      int
}

func NewStdFileLogger(plain bool, level int) (*Logger, error) {
	return NewLogger(true, true, plain, level)
}

func NewStdLogger(plain bool, level int) (*Logger, error) {
	return NewLogger(true, false, plain, level)
}

func NewFileLogger(plain bool, level int) (*Logger, error) {
	return NewLogger(false, true, plain, level)
}

func NewLogger(stdLog, fileLog, plain bool, level int) (*Logger, error) {
	l := &Logger{
		StdLog:  stdLog,
		FileLog: fileLog,
		Level:   level,
		Plain:   plain,
	}
	err := l.init(level)
	return l, err
}

func (l *Logger) init(level int) (err error) {
	multiWriter := io.MultiWriter()
	if l.StdLog {
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
	var flag int
	if !l.Plain {
		flag = log.Ltime | log.Ldate
	}

	l.BufferData = new(bytes.Buffer)
	multiWriter = io.MultiWriter(multiWriter, l.BufferData)

	if LevelFatal <= level {
		l.Fatal = log.New(multiWriter, "", flag)
	} else {
		l.Fatal = log.New(io.Discard, "", flag)
	}
	if LevelError <= level {
		l.Error = log.New(multiWriter, "", flag)
	} else {
		l.Error = log.New(io.Discard, "", flag)
	}
	if LevelWarning <= level {
		l.Warning = log.New(multiWriter, "", flag)
	} else {
		l.Warning = log.New(io.Discard, "", flag)
	}
	if LevelInfo <= level {
		l.Info = log.New(multiWriter, "", flag)
	} else {
		l.Info = log.New(io.Discard, "", flag)
	}
	if LevelDebug <= level {
		l.Debug = log.New(multiWriter, "", flag)
	} else {
		l.Debug = log.New(io.Discard, "", flag)
	}
	if LevelTrace <= level {
		l.Trace = log.New(multiWriter, "", flag)
	} else {
		l.Trace = log.New(io.Discard, "", flag)
	}
	return
}

func (logger *Logger) GetBuffer() (log string) {
	log = string(logger.BufferData.Bytes())
	return
}
