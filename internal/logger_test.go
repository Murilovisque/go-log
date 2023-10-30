package logs

import (
	"strings"
	"sync"
	"testing"
)

var (
	logWriter = logWriterTest{
		mux:          sync.Mutex{},
		lines:        []string{},
		persistLines: false,
	}
	sl = SimpleLogger{LevelSelected: LogDebugMode}
)

func TestShouldLogSimpleMessage(t *testing.T) {
	setup()
	sl.Info("teste")
	logWriter.assertLogMessage(t, "INFO * teste\n")
	sl.Error("teste")
	logWriter.assertLogMessage(t, "ERROR * teste\n")
}

func TestShouldLogSimpleMessageConcurrency(t *testing.T) {
	setup()
	logWriter.persistLines = true
	poolSize := 20
	chValues := make(chan int)
	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < poolSize; i++ {
		go func() {
			for v := range chValues {
				sl.Info(v)
			}
			wg.Done()
		}()
	}
	totalLines := 50000
	for i := 0; i < totalLines; i++ {
		v := i
		chValues <- v
	}
	close(chValues)
	wg.Wait()
	if len(logWriter.lines) != totalLines {
		t.Fatalf("Expected %d lines, but %d", totalLines, len(logWriter.lines))
	}
	uniques := make(map[string]int)
	for _, l := range logWriter.lines {
		uniques[l] = 1
	}
	if len(logWriter.lines) != len(uniques) {
		t.Fatalf("There are duplicated values")
	}
}

func TestShouldLogFormattedMessage(t *testing.T) {
	setup()
	sl.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO * teste txt 10\n")
	sl.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR * teste txt 10\n")
}

func TestShouldLogSimpleMessageWithFixedFields(t *testing.T) {
	setup(FieldValue{"reqid", "1"}, FieldValue{"idtperson", "2"})
	sl.Info("teste")
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste\n")
	sl.Error("teste")
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste\n")
}

func TestShouldLogFormattedMessageWithFixedFields(t *testing.T) {
	setup(FieldValue{"reqid", "1"}, FieldValue{"idtperson", "2"})
	sl.Infof("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "INFO [reqid: 1] [idtperson: 2] * teste txt 10\n")
	sl.Errorf("teste %s %d", "txt", 10)
	logWriter.assertLogMessage(t, "ERROR [reqid: 1] [idtperson: 2] * teste txt 10\n")
}

func setup(fixedValues ...FieldValue) {
	logWriter.lines = []string{}
	sl = SimpleLogger{FieldsValues: fixedValues, LevelSelected: LogDebugMode}
	sl.SetWriter(&logWriter)
	sl.Init()
}

type logWriterTest struct {
	mux          sync.Mutex
	lines        []string
	persistLines bool
}

func (w *logWriterTest) Write(p []byte) (n int, err error) {
	w.mux.Lock()
	w.lines = append(w.lines, string(p))
	if !w.persistLines {
		w.lines = w.lines[len(w.lines)-1 : len(w.lines)]
	}
	w.mux.Unlock()
	return len(p), nil
}

func (w *logWriterTest) assertLogMessage(t *testing.T, m string) {
	lastLog := w.lines[len(w.lines)-1]
	if !strings.HasSuffix(lastLog, m) {
		t.Fatalf("Expected '%s', but '%s' was logged\n", m, lastLog)
	}
}
