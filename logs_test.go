package logs

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSetupAndRotateLogFilePerDay(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
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
		t.Fatalf("%s failed, folder should 2 files, but it has %d", t.Name(), len(files))
	} else {
		for i, txt := range txts {
			if content, err := fileContent(dir, files[i]); err != nil || !strings.Contains(content, txt) {
				t.Fatalf("%s failed, folder expected %s, but received %s. Error: %v", txt, t.Name(), content, err)
			}
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
