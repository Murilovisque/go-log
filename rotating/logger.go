package rotating

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/Murilovisque/logs"
)

type TimeRotatingScheme int

const (
	PerDay TimeRotatingScheme = iota
	PerHour
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
	file                  *os.File
	mux                   sync.Mutex
	amountOfFilesToRetain int
}

var (
	ErrInvalidAmountOfFilesToRetain = errors.New("amount of files to retain is less than zero")
)

func NewTimeRotatingLogger(filename string, rotatingScheme TimeRotatingScheme, amountOfFilesToRetain int) (*TimeRotatingLogger, error) {
	if amountOfFilesToRetain < 1 {
		return nil, ErrInvalidAmountOfFilesToRetain
	}
	newFilename := buildFilenameWithTimeExtension(time.Now(), filename, rotatingScheme)
	f, err := os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	t := TimeRotatingLogger{rotatingScheme: rotatingScheme, filename: filename, file: f}
	go rotatingFile(&t)
	return &t, nil
}

func (trl *TimeRotatingLogger) Write(p []byte) (int, error) {
	trl.mux.Lock()
	n, err := trl.file.Write(p)
	trl.mux.Unlock()
	return n, err
}

func durationUntilNextRotating(moment time.Time, rotatingScheme TimeRotatingScheme) time.Duration {
	nextRotatingTime := moment.Add(rotatingScheme.rotatingInterval()).Truncate(rotatingScheme.rotatingInterval())
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
	filenameExt := path.Ext(trl.filename)
	filenameWithoutExt := trl.filename[:len(trl.filename)-len(filenameExt)]
	regexPattern := fmt.Sprintf("^%s-(%s)%s$", filenameWithoutExt, trl.rotatingScheme.timeExtensionRegex(), filenameExt)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return false
	}
	matchGroups := regex.FindStringSubmatch(filenameToCheck)
	if len(matchGroups) == 0 {
		return false
	}
	fileTime, err := time.Parse(trl.rotatingScheme.timeExtensionFormat(), matchGroups[len(matchGroups)-1])
	if err != nil {
		return false
	}
	return fileTime.Before(lastFileTime)
}

func removeOldFiles(moment time.Time, trl *TimeRotatingLogger) {
	filenameExt := path.Ext(trl.filename)
	filenameWithoutExtGlob := trl.filename[:len(trl.filename)-len(filenameExt)] + "*"
	dirEntries, err := filepath.Glob(filenameWithoutExtGlob)
	if err != nil {
		logs.Errorf("Glob %s failed. Is was not possible to remove old files - Error: %s", filenameWithoutExtGlob, err)
	} else {
		lastFileTime := lastFileTimeToRetain(moment, trl)
		for _, dirEntry := range dirEntries {
			if mustFileBeRemoved(lastFileTime, dirEntry, trl) {
				os.Remove(dirEntry)
			}
		}
	}
}

func rotatingFile(trl *TimeRotatingLogger) {
	next := durationUntilNextRotating(time.Now(), trl.rotatingScheme)
	tick := time.NewTicker(next)
	for {
		moment := <-tick.C
		newFilename := buildFilenameWithTimeExtension(moment, trl.filename, trl.rotatingScheme)
		f, err := os.OpenFile(newFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logs.Errorf("It was not possible rotate to file %s - Error: %s", newFilename, err)
		} else {
			trl.mux.Lock()
			trl.file.Close()
			trl.file = f
			trl.mux.Unlock()
		}
		removeOldFiles(moment, trl)
		next = durationUntilNextRotating(time.Now(), trl.rotatingScheme)
		tick.Reset(next)
	}
}
