package logs

import (
	"errors"
	"strings"

	logs "github.com/Murilovisque/logs/v3/internal"
	"github.com/Murilovisque/logs/v3/internal/rotating"
)

const (
	RotatingSchemaPerDay  = rotating.PerDay
	RotatingSchemaPerHour = rotating.PerHour
)

var (
	errTimeRotatingSchemeConversion = errors.New("time rotationg scheme conversion failed")
)

func InitWithRotatingLogFile(level logs.LoggerLevelMode, filename string, rotatingScheme rotating.TimeRotatingScheme, amountOfFilesToRetain int, compressOldFiles bool, fixedValues ...logs.FieldValue) error {
	l, err := rotating.NewTimeRotatingLogger(level, filename, rotatingScheme, amountOfFilesToRetain, compressOldFiles, fixedValues...)
	if err != nil {
		return err
	}
	return initGlobalLogger(level, l)
}

func StringToTimeRotatingScheme(s string) (rotating.TimeRotatingScheme, error) {
	s = strings.ToUpper(s)
	switch s {
	case strings.ToUpper(string(rotating.PerDay)):
		return rotating.PerDay, nil
	case strings.ToUpper(string(rotating.PerHour)):
		return rotating.PerHour, nil
	default:
		return "", errTimeRotatingSchemeConversion
	}
}
