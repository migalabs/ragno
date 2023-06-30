package crawler

import (
	"github.com/sirupsen/logrus"
)

type Llvl string

var (
	Trace  Llvl = "trace"
	Debug  Llvl = "debug"
	Info   Llvl = "info"
	Warn   Llvl = "warn"
	ErrorL Llvl = "error"
)

func ParseLogLevel(level string) logrus.Level {
	var lvl logrus.Level
	switch Llvl(level) {
	case Trace:
		lvl = logrus.TraceLevel
	case Debug:
		lvl = logrus.DebugLevel
	case Info:
		lvl = logrus.InfoLevel
	case Warn:
		lvl = logrus.WarnLevel
	case ErrorL:
		lvl = logrus.ErrorLevel
	default:
		lvl = logrus.InfoLevel
	}
	return lvl
}
