package logger

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("gesclient")

func nillable(args []interface{}) []interface{} {
	out := make([]interface{}, len(args))
	for i, a := range args {
		switch a.(type) {
		case *int:
			if a == nil {
				out[i] = "<nil>"
			} else {
				out[i] = *(a.(*int))
			}
		case *int32:
			if a == nil {
				out[i] = "<nil>"
			} else {
				out[i] = *(a.(*int32))
			}
		case *int64:
			if a == nil {
				out[i] = "<nil>"
			} else {
				out[i] = *(a.(*int64))
			}
		default:
			out[i] = a
		}
	}
	return out
}

func Fatal(args ...interface{}) {
	if len(args) > 0 {
		log.Fatal(nillable(args)...)
	} else {
		log.Fatal()
	}
}

func Fatalf(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Fatalf(format, nillable(args)...)
	} else {
		log.Fatalf(format)
	}
}

func Panic(args ...interface{}) {
	if len(args) > 0 {
		log.Panic(nillable(args)...)
	} else {
		log.Panic()
	}
}

func Panicf(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Panicf(format, nillable(args)...)
	} else {
		log.Panicf(format)
	}
}

func Critical(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Critical(nillable(args)...)
	} else {
		log.Critical(format)
	}
}

func Error(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Error(nillable(args)...)
	} else {
		log.Error(format)
	}
}

func Errorf(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Errorf(format, nillable(args)...)
	} else {
		log.Errorf(format)
	}
}

func Warning(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Warning(nillable(args)...)
	} else {
		log.Warning(format)
	}
}

func Warningf(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Warningf(format, nillable(args)...)
	} else {
		log.Warningf(format)
	}
}

func Notice(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Notice(nillable(args)...)
	} else {
		log.Notice(format)
	}
}

func Noticef(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Noticef(format, nillable(args)...)
	} else {
		log.Noticef(format)
	}
}

func Info(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Info(nillable(args)...)
	} else {
		log.Info(format)
	}
}

func Infof(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Infof(format, nillable(args)...)
	} else {
		log.Infof(format)
	}
}

func Debug(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Debug(nillable(args)...)
	} else {
		log.Debug(format)
	}
}

func Debugf(format string, args ...interface{}) {
	if len(args) > 0 {
		log.Debugf(format, nillable(args)...)
	} else {
		log.Debugf(format)
	}
}
