package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"

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

// Message structure for async logging
type logMessage struct {
	level   int
	message string
}

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
	// Adding fields for async logging
	async    bool
	msgChan  chan logMessage
	wg       sync.WaitGroup
	stopChan chan struct{}
}

func NewLogger() *Logger {
	return &Logger{
		Level:    LevelInfo,
		Flag:     0,
		Instant:  true,
		async:    true,
		msgChan:  make(chan logMessage, 1000),
		stopChan: make(chan struct{}),
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

// Set to async mode (default)
func (l *Logger) SetAsync() *Logger {
	l.async = true
	l.msgChan = make(chan logMessage, 1000) // Buffer size can be adjusted
	l.stopChan = make(chan struct{})
	return l
}

// Set to sync mode
func (l *Logger) SetSync() *Logger {
	l.async = false
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

	// Start async processing goroutine
	if l.async && l.msgChan != nil {
		l.wg.Add(1)
		go l.processLogs()
	}

	return l
}

// Process logs in background goroutine
func (l *Logger) processLogs() {
	defer l.wg.Done()

	// Create actual writers for logs
	multiWriter := io.MultiWriter()
	if l.Std != nil {
		multiWriter = io.MultiWriter(multiWriter, l.Std)
	}
	if l.File != nil {
		multiWriter = io.MultiWriter(multiWriter, l.File)
	}

	errorLogger := log.New(multiWriter, "", l.Flag)
	infoLogger := log.New(multiWriter, "", l.Flag)
	debugLogger := log.New(multiWriter, "", l.Flag)

	for {
		select {
		case msg := <-l.msgChan:
			switch msg.level {
			case LevelError:
				if LevelError <= l.Level {
					errorLogger.Print(msg.message)
				}
			case LevelInfo:
				if LevelInfo <= l.Level {
					infoLogger.Print(msg.message)
				}
			case LevelDebug:
				if LevelDebug <= l.Level {
					debugLogger.Print(msg.message)
				}
			}
		case <-l.stopChan:
			return
		}
	}
}

// Modified methods to support async writing
func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.async && l.msgChan != nil {
		select {
		case l.msgChan <- logMessage{level: LevelDebug, message: fmt.Sprintf(format, v...)}:
		default:
			// Queue full, write directly
			l.Debug.Printf(format, v...)
		}
	} else {
		l.Debug.Printf(format, v...)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.async && l.msgChan != nil {
		select {
		case l.msgChan <- logMessage{level: LevelInfo, message: fmt.Sprintf(format, v...)}:
		default:
			// Queue full, write directly
			l.Info.Printf(format, v...)
		}
	} else {
		l.Info.Printf(format, v...)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.async && l.msgChan != nil {
		select {
		case l.msgChan <- logMessage{level: LevelError, message: fmt.Sprintf(format, v...)}:
		default:
			// Queue full, write directly
			l.Error.Printf(format, v...)
		}
	} else {
		l.Error.Printf(format, v...)
	}
}

// Close logger and wait for all logs to be written
func (l *Logger) Close() {
	if l.async && l.stopChan != nil {
		close(l.stopChan)
		l.wg.Wait()
	}
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
