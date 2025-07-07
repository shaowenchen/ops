package log

import (
	"bytes"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/shaowenchen/ops/pkg/constants"
	"github.com/shaowenchen/ops/pkg/utils"
)

var Std *os.File = os.Stdout
var File *os.File = nil

const (
	LevelError = iota
	LevelInfo
	LevelDebug
)

func getLogVerbose(level string) int {
	levelInit, err := strconv.Atoi(level)
	if err == nil {
		return levelInit
	}
	switch level {
	case "error":
		return LevelError
	case "info":
		return LevelInfo
	case "debug":
		return LevelDebug
	default:
		return LevelInfo
	}
}

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

type Logger struct {
	Buffer  *bytes.Buffer
	Flag    int
	Level   int
	Instant bool
	Std     *os.File
	File    *os.File
	Debug   *log.Logger
	Info    *log.Logger
	Error   *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Level:   LevelInfo,
		Flag:    0,
		Instant: true,
	}
}

func (l *Logger) SetStd() *Logger {
	l.Std = Std
	return l
}

func (l *Logger) SetFile() *Logger {
	l.File = File
	return l
}

func (l *Logger) SetVerbose(level string) *Logger {
	l.Level = getLogVerbose(level)
	return l
}

func (l *Logger) SetFlag() *Logger {
	l.Flag = log.Ltime | log.Ldate
	return l
}

func (l *Logger) WaitFlush() *Logger {
	l.Instant = false
	return l
}

func (l *Logger) Build() *Logger {
	multiWriter := io.MultiWriter()

	if l.Instant {
		if l.Std != nil {
			multiWriter = io.MultiWriter(multiWriter, l.Std)
		}
		if l.File != nil {
			multiWriter = io.MultiWriter(multiWriter, l.File)
		}
	} else {
		if l.Buffer == nil {
			l.Buffer = bytes.NewBuffer([]byte{})
		}
		multiWriter = io.MultiWriter(multiWriter, l.Buffer)
	}

	if LevelError <= l.Level {
		l.Error = log.New(multiWriter, "", l.Flag)
	} else {
		l.Error = log.New(io.Discard, "", l.Flag)
	}
	if LevelInfo <= l.Level {
		l.Info = log.New(multiWriter, "", l.Flag)
	} else {
		l.Info = log.New(io.Discard, "", l.Flag)
	}
	if LevelDebug <= l.Level {
		l.Debug = log.New(multiWriter, "", l.Flag)
	} else {
		l.Debug = log.New(io.Discard, "", l.Flag)
	}
	return l
}

func (l *Logger) Flush() string {
	multiWriter := io.MultiWriter()
	if l.Std != nil {
		multiWriter = io.MultiWriter(multiWriter, l.Std)
	}
	if l.File != nil {
		multiWriter = io.MultiWriter(multiWriter, l.File)
	}
	content := l.Buffer.String()
	io.Copy(multiWriter, l.Buffer)
	defer l.Buffer.Reset()
	return content
}
