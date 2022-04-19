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
	"strings"
	"sync"
	"time"

	logs "github.com/Murilovisque/logs/v3/internal"
	"github.com/Murilovisque/logs/v3/internal/compressor"
)

type TimeRotatingScheme string

const (
	PerDay            TimeRotatingScheme = "perDay"
	PerHour           TimeRotatingScheme = "perHour"
	zipExtension                         = ".zip"
	zipExtensionRegex                    = "\\.zip"
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

func (trs TimeRotatingScheme) nextTruncatedTimeAfter(t time.Time) time.Time {
	switch trs {
	case PerDay:
		t = t.AddDate(0, 0, 1)
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
	currentLogFilename    string
	file                  io.Writer
	mux                   sync.Mutex
	amountOfFilesToRetain int
	compressOldFiles      bool
	closeSignalListener   chan int
	closedListener        chan int
	closed                bool
	logs.SimpleLogger
}

func NewTimeRotatingLogger(level logs.LoggerLevelMode, filename string, rotatingScheme TimeRotatingScheme, amountOfFilesToRetain int, compressOldFiles bool, fixedValues ...logs.FieldValue) (*TimeRotatingLogger, error) {
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
		currentLogFilename:    newFilename,
		file:                  f,
		closeSignalListener:   make(chan int),
		closedListener:        make(chan int, 1),
		amountOfFilesToRetain: amountOfFilesToRetain,
		compressOldFiles:      compressOldFiles,
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
	if !trl.closed {
		trl.closed = true
		trl.closeSignalListener <- 1
		moment := trl.rotatingScheme.nowTruncated()
		removeOldFiles(moment, trl)
		trl.mux.Lock()
		trl.file.(*os.File).Sync()
		trl.file.(*os.File).Close()
		trl.file = os.Stderr
		trl.mux.Unlock()
	}
	<-trl.closedListener
}

func durationUntilNextRotating(moment time.Time, rotatingScheme TimeRotatingScheme) time.Duration {
	nextRotatingTime := rotatingScheme.nextTruncatedTimeAfter(moment)
	nextDuration := nextRotatingTime.Sub(moment)
	if nextDuration < 1 {
		return 1
	}
	return nextDuration
}

func buildFilenameWithTimeExtension(moment time.Time, filename string, rotatingScheme TimeRotatingScheme) string {
	filenameExt := getFilenameExt(filename, true)
	filenameWithoutExt := filename[:len(filename)-len(filenameExt)]
	return fmt.Sprintf("%s-%s%s", filenameWithoutExt, moment.Format(rotatingScheme.timeExtensionFormat()), filenameExt)
}

func lastFileTimeToRetain(moment time.Time, trl *TimeRotatingLogger) time.Time {
	return moment.Add(trl.rotatingScheme.rotatingInterval() * time.Duration(trl.amountOfFilesToRetain) * -1)
}

func mustFileBeRemoved(lastFileTime time.Time, filenameToCheck string, trl *TimeRotatingLogger) bool {
	filenameEscaped := regexp.QuoteMeta(trl.filename)
	filenameExt := getFilenameExt(filenameEscaped, false)
	filenameWithoutExt := getFilenameWithoutExt(filenameEscaped)
	regexPattern := fmt.Sprintf("^%s-(%s)%s(%s)?$", filenameWithoutExt, trl.rotatingScheme.timeExtensionRegex(), filenameExt, zipExtensionRegex)
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		trl.Errorf("Error to generate the regex pattern to remove old files %v", err)
		return false
	}
	matchGroups := regex.FindStringSubmatch(filenameToCheck)
	if len(matchGroups) < 2 {
		return false
	}
	fileTime, err := time.ParseInLocation(trl.rotatingScheme.timeExtensionFormat(), matchGroups[1], lastFileTime.Location())
	if err != nil {
		return false
	}
	return fileTime.Before(lastFileTime)
}

func getFilenameWithoutExt(filename string) string {
	filenameExt := getFilenameExt(filename, true)
	return filename[:len(filename)-len(filenameExt)]
}

func getFilenameGlobWithoutExt(filename string) string {
	filenameExt := getFilenameExt(filename, true)
	return filename[:len(filename)-len(filenameExt)] + "*"
}

func getFilenameExt(filename string, includeCompressExtension bool) string {
	filenameSize := len(filename)
	if strings.HasSuffix(filename, zipExtension) {
		ext := path.Ext(filename[:filenameSize-len(zipExtension)])
		if includeCompressExtension {
			return ext + zipExtension
		}
		return ext
	}
	return path.Ext(filename)
}

func removeOldFiles(moment time.Time, trl *TimeRotatingLogger) {
	filenameWithoutExtGlob := getFilenameGlobWithoutExt(trl.filename)
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
				oldLogFilename := trl.currentLogFilename
				trl.mux.Lock()
				trl.file.(*os.File).Sync()
				trl.file.(*os.File).Close()
				trl.currentLogFilename = newFilename
				trl.file = f
				trl.mux.Unlock()
				if trl.compressOldFiles {
					err := compressor.ComprimirArquivo(oldLogFilename)
					if err != nil {
						trl.Errorf("It was not possible compress the file %s - Error: %s", oldLogFilename, err)
					} else {
						os.Remove(oldLogFilename)
					}
				}
				trl.Debugf("Log rotated to new file: %s", newFilename)
			}
			removeOldFiles(moment, trl)
			next = durationUntilNextRotating(time.Now(), trl.rotatingScheme)
			tick.Reset(next)
			trl.Debugf("Log rotating operation finished, next will be at %v", next)
		case <-trl.closeSignalListener:
			trl.closedListener <- 1
			return
		}
	}
}
