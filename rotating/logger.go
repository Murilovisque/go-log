package rotating

import (
	"fmt"
	"os"
	"path"
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

func (trs TimeRotatingScheme) timeExtension() string {
	switch trs {
	case PerDay:
		return time.Now().Format("20060102")
	case PerHour:
		return time.Now().Format("20060102-15")
	default:
		panic("Not implemented")
	}
}

type TimeRotatingLogger struct {
	rotatingScheme TimeRotatingScheme
	filename       string
	file           *os.File
	mux            sync.Mutex
}

func NewTimeRotatingLogger(filename string, rotatingScheme TimeRotatingScheme) (*TimeRotatingLogger, error) {
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
	return fmt.Sprintf("%s-%s%s", filenameWithoutExt, moment.Format(rotatingScheme.timeExtension()), filenameExt)
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
		next = durationUntilNextRotating(time.Now(), trl.rotatingScheme)
		tick.Reset(next)
	}
}
