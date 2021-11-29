package rotating

import (
	"testing"

	"github.com/Murilovisque/logs/v2/internal/rotating"
)

func TestShouldConvertStringToTimeScheme(t *testing.T) {
	s, err := StringToTimeRotatingScheme("perDay")
	if err != nil || s != rotating.PerDay {
		t.Fatal(err, s)
	}
	s, err = StringToTimeRotatingScheme("PERDAY")
	if err != nil || s != rotating.PerDay {
		t.Fatal(err, s)
	}

	s, err = StringToTimeRotatingScheme("perHour")
	if err != nil || s != rotating.PerHour {
		t.Fatal(err, s)
	}
	s, err = StringToTimeRotatingScheme("PERHOUR")
	if err != nil || s != rotating.PerHour {
		t.Fatal(err, s)
	}

}
