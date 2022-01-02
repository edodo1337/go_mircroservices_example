package storage

import "github.com/sirupsen/logrus"

func parseLogLevel(level string) logrus.Level {
	levelsMap := map[string]logrus.Level{
		"PANIC": logrus.PanicLevel,
		"FATAL": logrus.FatalLevel,
		"ERROR": logrus.ErrorLevel,
		"WARN":  logrus.WarnLevel,
		"INFO":  logrus.InfoLevel,
		"DEBUG": logrus.DebugLevel,
		"TRACE": logrus.TraceLevel,
	}

	value, exists := levelsMap[level]
	if !exists {
		return logrus.WarnLevel
	}

	return value
}
