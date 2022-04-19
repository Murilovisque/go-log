package rotating

import (
	"runtime"
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
		{"/varl/log/teste.log", "/varl/log/teste-20121207.log", PerDay},
		{"/varl/log/teste", "/varl/log/teste-20121207", PerDay},
		{"/varl/log/", "/varl/log/-20121207", PerDay},
		{"/varl/log/teste.log", "/varl/log/teste-20121207-06.log", PerHour},
		{"/varl/log/teste", "/varl/log/teste-20121207-06", PerHour},
		{"/varl/log/", "/varl/log/-20121207-06", PerHour},
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
		so           string
	}{
		{"/varl/log/teste.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"/varl/log/teste", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"/varl/log/teste.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121207.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"/varl/log/teste-20121208.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207.log", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121208.log", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},

		{"/varl/log/teste.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"/varl/log/teste.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/teste.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121207.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"/varl/log/teste-20121208.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207.log.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121208.log.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},

		{"/varl/log/teste-20121207-06.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121207-07.log", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-06.log", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-07.log", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121206.log", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206.log", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/teste-20121207-05.log", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121206-07.log", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-05.log", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-06.log", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-07.log", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},

		{"/varl/log/teste-20121207-06.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121207-07.log.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-06.log.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-07.log.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121206.log.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206.log.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/teste-20121207-05.log.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121206-07.log.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-05.log.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-06.log.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-07.log.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste.log", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},

		{"/varl/log/teste-20121207-06", false, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121207-07", false, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-06", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-07", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121206", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/teste-20121207-05", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121206-07", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-05", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-06", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-07", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},

		{"/varl/log/teste-20121207-06.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121207-07.zip", false, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-06.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121207-07.zip", false, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"/varl/log/teste-20121206.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerDay}, lastFileTimePerDay, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerDay}, lastFileTimePerDay, "windows"},
		{"/varl/log/teste-20121207-05.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"/varl/log/teste-20121206-07.zip", true, &TimeRotatingLogger{filename: "/varl/log/teste", rotatingScheme: PerHour}, lastFileTimePerHour, "linux"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-05.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-06.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
		{"C:\\Test.Legal\\Temp\\teste-20121206-07.zip", true, &TimeRotatingLogger{filename: "C:\\Test.Legal\\Temp\\teste", rotatingScheme: PerHour}, lastFileTimePerHour, "windows"},
	}
	os := runtime.GOOS
	for i, test := range tests {
		if !strings.HasPrefix(test.so, os) {
			continue
		}
		must := mustFileBeRemoved(test.lastFileTime, test.vl, test.trl)
		if must != test.exp {
			t.Fatal(i, test)
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

func TestGetFilenameGlobWithoutExt(t *testing.T) {
	tests := []struct {
		vl       string
		expected string
	}{
		{"/var/log/teste.log", "/var/log/teste*"},
		{"/var/log/teste-.log", "/var/log/teste-*"},
		{"/var/log/teste-23.log", "/var/log/teste-23*"},
		{"/var/log/teste.log.zip", "/var/log/teste*"},
		{"teste.log", "teste*"},
		{"/var/log/teste", "/var/log/teste*"},
		{"teste", "teste*"},
		{"teste.log.zip", "teste*"},
	}
	for _, test := range tests {
		vl := getFilenameGlobWithoutExt(test.vl)
		if vl != test.expected {
			t.Fatalf("Expected %s, but received %s", test.expected, vl)
		}
	}
}

func TestGetFilenameExt(t *testing.T) {
	tests := []struct {
		vl       string
		expected string
	}{
		{"/var/log/teste.log", ".log"},
		{"/var/log/teste-.log", ".log"},
		{"/var/log/teste-23.log", ".log"},
		{"teste.log", ".log"},
		{"/var/log/teste", ""},
		{"teste", ""},
		{"teste.log.zip", ".log.zip"},
		{"/var/log/teste.log.zip", ".log.zip"},
	}
	for _, test := range tests {
		vl := getFilenameExt(test.vl, true)
		if vl != test.expected {
			t.Fatalf("Expected %s, but received %s", test.expected, vl)
		}
	}
}

func TestGetFilenameWithoutExt(t *testing.T) {
	tests := []struct {
		vl       string
		expected string
	}{
		{"/var/log/teste.log", "/var/log/teste"},
		{"/var/log/teste-.log", "/var/log/teste-"},
		{"/var/log/teste-23.log", "/var/log/teste-23"},
		{"/var/log/teste.log.zip", "/var/log/teste"},
		{"teste.log", "teste"},
		{"/var/log/teste", "/var/log/teste"},
		{"teste", "teste"},
		{"teste.log.zip", "teste"},
	}
	for _, test := range tests {
		vl := getFilenameWithoutExt(test.vl)
		if vl != test.expected {
			t.Fatalf("Expected %s, but received %s", test.expected, vl)
		}
	}
}
