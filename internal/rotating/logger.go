package rotating

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	logs "github.com/Murilovisque/logs/v2/internal"
)

type TimeRotatingScheme string

const (
	PerDay  TimeRotatingScheme = "perDay"
	PerHour TimeRotatingScheme = "perHour"
)

var (
	ErrInvalidAmountOfFilesToRetain = errors.New("amount of files to retain is less than zero")
)

func (trs TimeRotatingScheme) rotatingInterval() time.Duration {
	switch trs {
	case PerDay:
		return time.Hour * 24
	case PerHour:
		return time.Hour
	default:
		panic("Not implemented")
	}
}

func (trs TimeRotatingScheme) nextTimeAfter(t time.Time) time.Time {
	switch trs {
	case PerDay:
		t = t.Add(time.Hour * 24)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case PerHour:
		t = t.Add(time.Hour)
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	default:
		panic("Not implemented")
	}
}

func (trs TimeRotatingScheme) nowTruncated() time.Time {
	t := time.Now()
	switch trs {
	case PerDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case PerHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	default:
		panic("Not implemented")
	}
}

func (trs TimeRotatingScheme) timeExtensionFormat() string {
	switch trs {
	case PerDay:
		return "20060102"
	case PerHour:
		return "20060102-15"
	default:
		panic("Not implemented")
	}
}

func (trs TimeRotatingScheme) timeExtensionRegex() string {
	switch trs {
	case PerDay:
		return "\\d{8}"
	case PerHour:
		return "\\d{8}-\\d{2}"
	default:
		panic("Not implemented")
	}
}

type TimeRotatingLogger struct {
	rotatingScheme        TimeRotatingScheme
	filename              string
	file                  io.Writer
	mux                   sync.Mutex
	amountOfFilesToRetain int
	closeListener         chan int
	closed                bool
	logs.SimpleLogger
}

func NewTimeRotatingLogger(level logs.LoggerLevelMode, filename string, rotatingScheme TimeRotatingScheme, amountOfFilesToRetain int, fixedValues ...logs.FieldValue) (*TimeRotatingLogger, error) {
	if amountOfFilesToRetain < 0 {
		return nil, ErrInvalidAmountOfFilesToRetain
	}
	newFilename := buildFilenameWithTimeExtension(time.Now(), filename, rotatingScheme)
	f, err := os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	t := TimeRotatingLogger{
		rotatingScheme:        rotatingScheme,
		filename:              filename,
		file:                  f,
		closeListener:         make(chan int),
		amountOfFilesToRetain: amountOfFilesToRetain,
		SimpleLogger:          logs.SimpleLogger{FieldsValues: fixedValues[:], LevelSelected: level},
	}
	return &t, nil
}

func (trl *TimeRotatingLogger) Init() {
	trl.SimpleLogger.Init()
	log.SetOutput(trl)
	go rotatingFile(trl)
}

func (trl *TimeRotatingLogger) Write(p []byte) (int, error) {
	trl.mux.Lock()
	n, err := trl.file.Write(p)
	trl.mux.Unlock()
	return n, err
}

func (trl *TimeRotatingLogger) SetWriter(writer io.Writer) {
	log.SetOutput(trl)
}

func (trl *TimeRotatingLogger) Close() {
	trl.mux.Lock()
	if !trl.closed {
		trl.closed = true
		trl.closeListener <- 1
		trl.file.(*os.File).Sync()
		trl.file.(*os.File).Close()
		trl.file = os.Stderr
	}
	trl.mux.Unlock()
}

func durationUntilNextRotating(moment time.Time, rotatingScheme TimeRotatingScheme) time.Duration {
	nextRotatingTime := rotatingScheme.nextTimeAfter(moment)
	nextDuration := nextRotatingTime.Sub(moment)
	if nextDuration < 1 {
		return 1
	}
	return nextDuration
}

func buildFilenameWithTimeExtension(moment time.Time, filename string, rotatingScheme TimeRotatingScheme) string {
	filenameExt := path.Ext(filename)
	filenameWithoutExt := filename[:len(filename)-len(filenameExt)]
	return fmt.Sprintf("%s-%s%s", filenameWithoutExt, moment.Format(rotatingScheme.timeExtensionFormat()), filenameExt)
}

func lastFileTimeToRetain(moment time.Time, trl *TimeRotatingLogger) time.Time {
	return moment.Add(trl.rotatingScheme.rotatingInterval() * time.Duration(trl.amountOfFilesToRetain) * -1)
}

func mustFileBeRemoved(lastFileTime time.Time, filenameToCheck string, trl *TimeRotatingLogger) bool {
	filenameEscaped := regexp.QuoteMeta(trl.filename)
	filenameExt := path.Ext(filenameEscaped)
	filenameWithoutExt := filenameEscaped[:len(filenameEscaped)-len(filenameExt)]
	regexPattern := fmt.Sprintf("^%s-(%s)%s$", filenameWithoutExt, trl.rotatingScheme.timeExtensionRegex(), filenameExt)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		trl.Errorf("Error to generate the regex pattern to remove old files %v", err)
		return false
	}
	matchGroups := regex.FindStringSubmatch(filenameToCheck)
	if len(matchGroups) == 0 {
		return false
	}
	fileTime, err := time.ParseInLocation(trl.rotatingScheme.timeExtensionFormat(), matchGroups[len(matchGroups)-1], lastFileTime.Location())
	if err != nil {
		return false
	}
	return fileTime.Before(lastFileTime)
}

func removeOldFiles(moment time.Time, trl *TimeRotatingLogger) {
	filenameExt := path.Ext(trl.filename)
	filenameWithoutExtGlob := trl.filename[:len(trl.filename)-len(filenameExt)] + "*"
	fileEntries, err := filepath.Glob(filenameWithoutExtGlob)
	if err != nil {
		trl.Errorf("Glob %s failed. Is was not possible to remove old files - Error: %s", filenameWithoutExtGlob, err)
	} else {
		lastFileTime := lastFileTimeToRetain(moment, trl)
		trl.Debugf("Last file moment to retain %v", lastFileTime)
		for _, filename := range fileEntries {
			if mustFileBeRemoved(lastFileTime, filename, trl) {
				err := os.Remove(filename)
				if err != nil {
					trl.Errorf("Is was not possible to remove the old file %s - Error: %s", filename, err)
				}
			}
		}
	}
}

func rotatingFile(trl *TimeRotatingLogger) {
	trl.Infof("Starting the log rotation: %v scheme", trl.rotatingScheme)
	next := durationUntilNextRotating(time.Now(), trl.rotatingScheme)
	trl.Debugf("Next log rotation will be at %v", next)
	tick := time.NewTicker(next)
	for {
		select {
		case <-tick.C:
			moment := trl.rotatingScheme.nowTruncated()
			trl.Debugf("Starting log rotating operation %v", moment)
			newFilename := buildFilenameWithTimeExtension(moment, trl.filename, trl.rotatingScheme)
			f, err := os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				trl.Errorf("It was not possible rotate to file %s - Error: %s", newFilename, err)
			} else {
				trl.mux.Lock()
				trl.file.(*os.File).Close()
				trl.file = f
				trl.mux.Unlock()
				trl.Debugf("Log rotated to new file: %s", newFilename)
			}
			removeOldFiles(moment, trl)
			next = durationUntilNextRotating(time.Now(), trl.rotatingScheme)
			trl.Debugf("Log rotating operation finished, next will be at %v", next)
			tick.Reset(next)
		case <-trl.closeListener:
			return
		}
	}
}
