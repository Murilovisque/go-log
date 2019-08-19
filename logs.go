package logs

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	logr            *logger
	shutdownChannel chan struct{}
	shutdownMode    = true
	bufPool         = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	waitIncreasedChannel chan struct{}
)

func init() {
	waitIncreasedChannel = make(chan struct{})
}

// Shutdown wait as logs complete
func Shutdown() {
	shutdownMode = true
	for {
		if len(logr.messagesChannel) == 0 {
			break
		}
	}
	close(shutdownChannel)
}

// LogMessageQueueSize returns the size of the logs message queue
func LogMessageQueueSize() int {
	if logr == nil {
		return 0
	}
	return cap(logr.messagesChannel)
}

// NumberOfMessagesQueued returns the amount of the logs message in the queue
func NumberOfMessagesQueued() int {
	if logr == nil {
		return 0
	}
	return len(logr.messagesChannel)
}

// SetupPerDay set the standard log to file and rotate it per day
func SetupPerDay(logPath string, logMessagesQueueSize int) error {
	nextDateLogFileFunc := func(now time.Time) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	}
	fileDateFormatFunc := func() string {
		return "2006-01-02"
	}
	return setup(nextDateLogFileFunc, fileDateFormatFunc, logPath, logMessagesQueueSize)
}

type logger struct {
	messagesChannel chan *bytes.Buffer
	logFile         *os.File
	logPath         string
	mux             sync.Mutex
}

func (l *logger) Write(p []byte) (int, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	if n, err := buf.Write(p); err != nil {
		buf.Reset()
		bufPool.Put(buf)
		return n, err
	}
	l.mux.Lock()
	defer l.mux.Unlock()
	select {
	case l.messagesChannel <- buf:
	default:
		messagesChannelIncreased := make(chan *bytes.Buffer, cap(l.messagesChannel)*2)
		for {
			if len(logr.messagesChannel) == 0 {
				break
			}
		}
		waitIncreasedChannel <- struct{}{}
		l.messagesChannel = messagesChannelIncreased
		waitIncreasedChannel <- struct{}{}
		l.messagesChannel <- buf
	}
	return len(p), nil
}

func setup(nextDateLogFileFunc func(now time.Time) time.Time, fileDateFormatFunc func() string, logPath string, logMessagesQueueSize int) error {
	shutdownMode = false
	shutdownChannel = make(chan struct{})
	filePathTest := fmt.Sprintf("%s%d", logPath, rand.Int())
	_, err := os.Create(filePathTest)
	if err != nil {
		return errors.New("It was not possible to log in " + logPath)
	}
	os.Remove(filePathTest)
	logr = &logger{
		messagesChannel: make(chan *bytes.Buffer, logMessagesQueueSize),
		logPath:         logPath,
	}
	openLogFile(nextDateLogFileFunc, fileDateFormatFunc)
	startMessagesLogger()
	log.SetOutput(logr)
	return nil
}

func openLogFile(nextDateLogFileFunc func(time.Time) time.Time, fileDateFormatFunc func() string) {
	if shutdownMode {
		return
	}
	now := time.Now()
	logPathWithDate := fmt.Sprintf("%s-%s.log", logr.logPath, now.Format(fileDateFormatFunc()))
	if _, err := os.Stat(logPathWithDate); os.IsNotExist(err) && logr.logFile != nil {
		oldFile := logr.logFile
		defer oldFile.Close()
	}
	file, err := os.OpenFile(logPathWithDate, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		logr.logFile = file
		time.AfterFunc(nextDateLogFileFunc(now).Sub(now), func() {
			log.Println("Rotating log file...")
			openLogFile(nextDateLogFileFunc, fileDateFormatFunc)
		})
	} else {
		panic(err)
	}
}

func startMessagesLogger() {
	go func() {
		for {
			select {
			case b := <-logr.messagesChannel:
				if _, err := b.WriteTo(logr.logFile); err != nil {
					panic(err)
				}
				b.Reset()
				bufPool.Put(b)
			case <-waitIncreasedChannel:
				<-waitIncreasedChannel
			case <-shutdownChannel:
				return
			}
		}
	}()
}
