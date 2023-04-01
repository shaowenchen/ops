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
	Trace      *log.Logger
	Debug      *log.Logger
	Info       *log.Logger
	Warning    *log.Logger
	Error      *log.Logger
	Fatal      *log.Logger
	BufferData *bytes.Buffer
	Level      int
}

func NewStdFileLogger(withTime bool, level int) (*Logger, error) {
	return NewLogger(true, true, withTime, level)
}

func NewStdLogger(withTime bool, level int) (*Logger, error) {
	return NewLogger(true, false, withTime, level)
}

func NewFileLogger(withTime bool, level int) (*Logger, error) {
	return NewLogger(false, true, withTime, level)
}

func NewLogger(stdLog, fileLog, withTime bool, level int) (*Logger, error) {
	var flag int
	if withTime {
		flag = log.Ltime | log.Ldate
	}
	l := &Logger{
		StdLog:  stdLog,
		FileLog: fileLog,
		Level:   level,
	}
	err := l.init(flag, level)
	return l, err
}

func (l *Logger) init(flag, level int) (err error) {
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

	l.BufferData = new(bytes.Buffer)
	multiWriter = io.MultiWriter(multiWriter, l.BufferData)

	if LevelFatal <= level {
		l.Fatal = log.New(multiWriter, "[FATAL]", flag)
	} else {
		l.Fatal = log.New(io.Discard, "[FATAL]", flag)
	}
	if LevelError <= level {
		l.Error = log.New(multiWriter, "[ERROR]", flag)
	} else {
		l.Error = log.New(io.Discard, "[ERROR]", flag)
	}
	if LevelWarning <= level {
		l.Warning = log.New(multiWriter, "[WARNING]", flag)
	} else {
		l.Warning = log.New(io.Discard, "[WARNING]", flag)
	}
	if LevelInfo <= level {
		l.Info = log.New(multiWriter, "[INFO]", flag)
	} else {
		l.Info = log.New(io.Discard, "[INFO]", flag)
	}
	if LevelDebug <= level {
		l.Debug = log.New(multiWriter, "[DEBUG]", flag)
	} else {
		l.Debug = log.New(io.Discard, "[DEBUG]", flag)
	}
	if LevelTrace <= level {
		l.Trace = log.New(multiWriter, "[TRACE]", flag)
	} else {
		l.Trace = log.New(io.Discard, "[TRACE]", flag)
	}
	return
}

func (logger *Logger) GetBuffer() (log string) {
	log = string(logger.BufferData.Bytes())
	return
}
