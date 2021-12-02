package logs

import (
	"errors"
	"strings"

	logs "github.com/Murilovisque/logs/v2/internal"
	"github.com/Murilovisque/logs/v2/internal/rotating"
)

const (
	RotatingSchemaPerDay = rotating.PerDay
	RotatingSchemaPerHour = rotating.PerHour
)

var (
	errTimeRotatingSchemeConversion = errors.New("Time rotationg scheme conversion failed")
)

func InitWithRotatingLogFile(level logs.LoggerLevelMode, filename string, rotatingScheme rotating.TimeRotatingScheme, amountOfFilesToRetain int, fixedValues ...logs.FieldValue) error {
	l, err := rotating.NewTimeRotatingLogger(filename, rotatingScheme, amountOfFilesToRetain, fixedValues...)
	if err != nil {
		return err
	}
	return initGlobalLogger(level, l)
}

func StringToTimeRotatingScheme(s string) (rotating.TimeRotatingScheme, error)  {
	s = strings.ToUpper(s)
	switch(s) {
	case strings.ToUpper(string(rotating.PerDay)):
		return rotating.PerDay, nil
	case strings.ToUpper(string(rotating.PerHour)):
		return rotating.PerHour, nil
	default:
		return "", errTimeRotatingSchemeConversion
	}
}

