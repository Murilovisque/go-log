package rotating

import (
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDurationUntilNextRotating(t *testing.T) {
	now, _ := time.Parse("2006 Jan 02 15:04:05", "2012 Dec 07 12:15:30")
	expectedPerDay := 11*time.Hour + 44*time.Minute + 30*time.Second
	vl := durationUntilNextRotating(now, PerDay)
	if vl != expectedPerDay {
		t.Fatal(vl)
	}
	expectedPerHour := 44*time.Minute + 30*time.Second
	vl = durationUntilNextRotating(now, PerHour)
	if vl != expectedPerHour {
		t.Fatal(vl)
	}
}

func TestBuildFilenameWithTimeExtension(t *testing.T) {
	now, _ := time.Parse("2006 Jan 02 15:04:05", "2012 Dec 07 06:15:30")
	tests := []struct {
		vl             string
		exp            string
		rotatingScheme TimeRotatingScheme
	}{
		{equalizePathSeparator("/varl/log/teste.log"), equalizePathSeparator("/varl/log/teste-20121207.log"), PerDay},
		{equalizePathSeparator("/varl/log/teste"), equalizePathSeparator("/varl/log/teste-20121207"), PerDay},
		{equalizePathSeparator("/varl/log/"), equalizePathSeparator("/varl/log/-20121207"), PerDay},
		{equalizePathSeparator("/varl/log/teste.log"), equalizePathSeparator("/varl/log/teste-20121207-06.log"), PerHour},
		{equalizePathSeparator("/varl/log/teste"), equalizePathSeparator("/varl/log/teste-20121207-06"), PerHour},
		{equalizePathSeparator("/varl/log/"), equalizePathSeparator("/varl/log/-20121207-06"), PerHour},
	}
	for _, test := range tests {
		f := buildFilenameWithTimeExtension(now, test.vl, test.rotatingScheme)
		if f != test.exp {
			t.Fatal(f)
		}
	}
}

func TestMustFileBeRemoved(t *testing.T) {
	lastFileTimePerDay, _ := time.Parse("2006 Jan 02", "2012 Dec 07")
	lastFileTimePerHour, _ := time.Parse("2006 Jan 02 15", "2012 Dec 07 06")
	tests := []struct {
		vl           string
		exp          bool
		trl          *TimeRotatingLogger
		lastFileTime time.Time
	}{
		{equalizePathSeparator("/varl/log/teste.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/teste"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/teste.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/teste"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/teste-20121207.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/teste-20121208.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/teste-20121207-06.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/teste-20121207-07.log"), false, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/teste-20121206.log"), true, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerDay}, lastFileTimePerDay},
		{equalizePathSeparator("/varl/log/teste-20121207-05.log"), true, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
		{equalizePathSeparator("/varl/log/teste-20121206-07.log"), true, &TimeRotatingLogger{filename: equalizePathSeparator("/varl/log/teste.log"), rotatingScheme: PerHour}, lastFileTimePerHour},
	}
	for _, test := range tests {
		must := mustFileBeRemoved(test.lastFileTime, test.vl, test.trl)
		if must != test.exp {
			t.Fatal(test)
		}
	}
}

func TestLastFileTimeToRetain(t *testing.T) {
	lastFileTimePerDay, _ := time.Parse("2006 Jan 02", "2012 Dec 07")
	lastFileTimePerHour, _ := time.Parse("2006 Jan 02 15", "2012 Dec 07 06")
	tests := []struct {
		vl  time.Time
		exp time.Time
		trl *TimeRotatingLogger
	}{
		{lastFileTimePerDay, time.Date(2012, 12, 7, 0, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerDay, amountOfFilesToRetain: 0}},
		{lastFileTimePerDay, time.Date(2012, 12, 6, 0, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerDay, amountOfFilesToRetain: 1}},
		{lastFileTimePerDay, time.Date(2012, 11, 27, 0, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerDay, amountOfFilesToRetain: 10}},
		{lastFileTimePerHour, time.Date(2012, 12, 7, 6, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerHour, amountOfFilesToRetain: 0}},
		{lastFileTimePerHour, time.Date(2012, 12, 7, 5, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerHour, amountOfFilesToRetain: 1}},
		{lastFileTimePerHour, time.Date(2012, 12, 7, 0, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerHour, amountOfFilesToRetain: 6}},
		{lastFileTimePerHour, time.Date(2012, 11, 6, 22, 0, 0, 0, time.Now().Location()), &TimeRotatingLogger{rotatingScheme: PerHour, amountOfFilesToRetain: 8}},
	}
	for _, test := range tests {
		last := lastFileTimeToRetain(test.vl, test.trl)
		if last.Equal(test.exp) {
			t.Fatal(last)
		}
	}
}

func equalizePathSeparator(p string) string {
    np := path.Join(strings.Split(p, "/")...)
    if strings.HasSuffix(p, "/") && !strings.HasSuffix(np, "/") {
        np += string(filepath.Separator)
    }
    return np
}
