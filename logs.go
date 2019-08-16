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
	shutdownMode    bool
	bufPool         = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

// Shutdown wait as logs complete
func Shutdown() {
	shutdownMode = true
	for {
		if len(logr.messagesChannel) == 0 {
			break
		}
	}
	shutdownChannel <- struct{}{}
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
	messagesChannel          chan *bytes.Buffer
	messagesChannelIncreased chan *bytes.Buffer
	logFile                  *os.File
	fatalFail                error
	logPath                  string
	mux                      sync.Mutex
}

func (l *logger) Write(p []byte) (int, error) {
	if l.fatalFail != nil {
		panic(l.fatalFail)
	}
	buf := bufPool.Get().(*bytes.Buffer)
	if n, err := buf.Write(p); err != nil {
		return n, err
	}
	l.mux.Lock()
	select {
	case l.messagesChannel <- buf:
	default:
		if l.messagesChannelIncreased == nil {
			l.messagesChannelIncreased = make(chan *bytes.Buffer, cap(l.messagesChannel)*2)
			go func() {
				close(logr.messagesChannel)
				for {
					if len(logr.messagesChannel) == 0 {
						break
					}
				}
				l.mux.Lock()
				l.messagesChannel = l.messagesChannelIncreased
				l.messagesChannelIncreased = nil
				l.mux.Unlock()
			}()
		}
		l.messagesChannelIncreased <- buf
	}
	l.mux.Unlock()
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
	if _, err := os.Stat(logPathWithDate); os.IsNotExist(err) {
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
		logr.fatalFail = err
	}
}

func startMessagesLogger() {
	go func() {
		for {
			select {
			case b := <-logr.messagesChannel:
				if _, err := b.WriteTo(logr.logFile); err != nil {
					logr.fatalFail = err
					return
				}
				b.Reset()
				bufPool.Put(b)
				break
			case <-shutdownChannel:
				return
			}
		}
	}()
}
