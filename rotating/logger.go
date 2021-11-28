package rotating

import (
	"errors"
	"strings"

	"github.com/Murilovisque/logs/v2/internal/rotating"
)

var (
	ErrTimeRotatingSchemeConversion = errors.New("Time rotationg scheme conversion failed")
)

func StringToTimeRotatingScheme(s string) (rotating.TimeRotatingScheme, error)  {
	s = strings.ToUpper(s)
	switch(s) {
	case rotating.PerDayConfig:
		return rotating.PerDay, nil
	case rotating.PerHourConfig:
		return rotating.PerHour, nil
	default:
		return -1, ErrTimeRotatingSchemeConversion
	}
}

