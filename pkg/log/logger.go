package log

import (
	"bytes"
	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
	"io"
	"log"
	"os"
)

var Std *os.File = os.Stdout
var File *os.File = nil

const (
	LevelError = iota
	LevelInfo
	LevelDebug
)

func init() {
	// init file logger
	err := utils.CreateDir(constants.GetOpsLogsDir())
	if err != nil {
		panic(err)
	}
	File, err = os.OpenFile(constants.GetOpsLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
}

func BuilderStdLogger(level int, plain bool, instant bool) *Logger {
	l := &Logger{
		Std:   Std,
		File:  nil,
		Level: level,
		Plain: plain,
	}
	return l.init(instant)
}

func BuilderFileLogger(level int, plain bool, instant bool) *Logger {
	l := &Logger{
		Std:   nil,
		File:  File,
		Level: level,
		Plain: plain,
	}
	return l.init(instant)
}

func BuilderStdFileLogger(level int, plain bool, instant bool) *Logger {
	l := &Logger{
		Std:   Std,
		File:  File,
		Level: level,
		Plain: plain,
	}
	return l.init(instant)
}

type Logger struct {
	Buffer *bytes.Buffer
	Plain  bool
	Level  int
	Std    *os.File
	File   *os.File
	Debug  *log.Logger
	Info   *log.Logger
	Error  *log.Logger
}

func (block *Logger) init(instant bool) *Logger {
	multiWriter := io.MultiWriter()

	var flag int
	if !block.Plain {
		flag = log.Ltime | log.Ldate
	}
	if instant {
		if block.Std != nil {
			multiWriter = io.MultiWriter(multiWriter, block.Std)
		}
		if block.File != nil {
			multiWriter = io.MultiWriter(multiWriter, block.File)
		}
	} else {
		if block.Buffer == nil {
			block.Buffer = bytes.NewBuffer([]byte{})
		}
		multiWriter = io.MultiWriter(multiWriter, block.Buffer)
	}

	if LevelError <= block.Level {
		block.Error = log.New(multiWriter, "", flag)
	} else {
		block.Error = log.New(io.Discard, "", flag)
	}
	if LevelInfo <= block.Level {
		block.Info = log.New(multiWriter, "", flag)
	} else {
		block.Info = log.New(io.Discard, "", flag)
	}
	if LevelDebug <= block.Level {
		block.Debug = log.New(multiWriter, "", flag)
	} else {
		block.Debug = log.New(io.Discard, "", flag)
	}
	return block
}

func (block *Logger) Flush() string {
	multiWriter := io.MultiWriter()
	if block.Std != nil {
		multiWriter = io.MultiWriter(multiWriter, block.Std)
	}
	if block.File != nil {
		multiWriter = io.MultiWriter(multiWriter, block.File)
	}
	io.Copy(multiWriter, block.Buffer)
	defer block.Buffer.Reset()
	return block.Buffer.String()
}
