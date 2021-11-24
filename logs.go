package logs

import logs "github.com/Murilovisque/logs/internal"


func InitWithLogFile(filename string, fixedValues ...logs.FieldValue) error {
	l, err := logs.NewLoggerWithLogFile(filename, fixedValues...)
	if err != nil {
		return err
	}
	logs.Init(l)
	return nil
}

