package logs

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIfSetupShutdownAndRotateLogFileWorks(t *testing.T) {
	dir := createTempDir(t)
	defer cleanup(dir, t)
	t.Log(dir)
	logFile := dir + string(os.PathSeparator) + "testfile"
	rotateAfterOneSecond := func(now time.Time) time.Time {
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second()+1, 0, time.Local)
	}
	rotateSecondLogFile := func() string {
		return "2006-01-02-15-04-5"
	}
	setup(rotateAfterOneSecond, rotateSecondLogFile, logFile, 5)
	txts := []string{"text-1", "text-2", "text-3", "text-4"}
	for i, txt := range txts {
		log.Println(txt)
		if i < len(txts)-1 {
			time.Sleep(time.Second)
		}
	}
	Shutdown()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(txts) {
		t.Fatalf("Folder should %d files, but it has %d", len(txts), len(files))
	}
	for i, txt := range txts {
		if content, err := fileContent(dir, files[i]); err != nil || !strings.Contains(content, txt) {
			t.Fatalf("Folder expected %s, but received %s. Error: %v", txt, content, err)
		}
	}
}

func TestIfTheLogMessageQueueWasIncreasedWhenThereWereManyLogsOperations(t *testing.T) {
	dir := createTempDir(t)
	defer cleanup(dir, t)
	t.Log(dir)
	logFile := dir + string(os.PathSeparator) + "testfile"
	err := SetupPerDay(logFile, 3)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	const expectedLines = 2000
	for i := 1; i <= expectedLines; i++ {
		go func(ind int) {
			wg.Add(1)
			log.Println(strings.Repeat("a", ind))
			wg.Done()
		}(i)
	}
	wg.Wait()
	Shutdown()
	if LogMessageQueueSize() <= 3 {
		t.Fatal("Log message queue size did not increased")
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("There should be one log file, but there are %d\n", len(files))
	}
	txt, err := fileContent(dir, files[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(txt, "\n")) != expectedLines+1 {
		t.Fatal("logs lost", LogMessageQueueSize(), len(strings.Split(txt, "\n")))
	}
}

func TestIfTheLogMessageWereLoggedInRightOrder(t *testing.T) {
	dir := createTempDir(t)
	defer cleanup(dir, t)
	t.Log(dir)
	logFile := dir + string(os.PathSeparator) + "testfile"
	err := SetupPerDay(logFile, 3)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	const expectedLines = 2000
	for i := 1; i <= expectedLines; i++ {
		log.Println(i)
	}
	wg.Wait()
	Shutdown()
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("There should be one log file, but there are %d\n", len(files))
	}
	txt, err := fileContent(dir, files[0])
	if err != nil {
		t.Fatal(err)
	}
	linesOfTxt := strings.Split(txt, "\n")
	if len(linesOfTxt) != expectedLines+1 {
		t.Fatal("logs lost", LogMessageQueueSize(), len(strings.Split(txt, "\n")))
	}
	for i := 0; i < expectedLines; i++ {
		if !strings.HasSuffix(linesOfTxt[i], strconv.Itoa(i+1)) {
			t.Fatalf("Invalid log order in line %d. Expected: %d - Received: %s", i+1, i+1, linesOfTxt[i])
		}
	}
}

func fileContent(dir string, fileInfo os.FileInfo) (string, error) {
	b, err := ioutil.ReadFile(dir + string(os.PathSeparator) + fileInfo.Name())
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func createTempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func cleanup(dir string, t *testing.T) {
	if !t.Failed() {
		os.RemoveAll(dir)
	}
}
