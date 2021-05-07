package rotating

import (
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
