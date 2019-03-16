package log

import (
	"fmt"
	"os"
	"time"
)

type Level int

const (
	CRITICAL Level = iota
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

var levels = [6]string {
	"CRITICAL",
	"ERROR",
	"WARNING",
	"NOTICE",
	"INFO",
	"DEBUG",
}

var level Level = ERROR

func SetLevel(l Level) {
	level = l
}

const (
	logFormat = "%s - %s - %s\n"
	timeFormat = "2006-01-02T15:04:05.999999"
)

func output(level Level, msg string) {
	_, err := fmt.Fprintf(os.Stderr, logFormat, time.Now().UTC().Format(timeFormat), levels[level], msg)
	if err != nil {
		panic(err)
	}
}

func Critical(msg string) {
	if level < CRITICAL {
		return
	}
	output(CRITICAL, msg)
}

func Criticalf(format string, v ...interface{}) {
	if level < CRITICAL {
		return
	}
	output(CRITICAL, fmt.Sprintf(format, v...))
}

func Error(msg string) {
	if level < ERROR {
		return
	}
	output(ERROR, msg)
}

func Errorf(format string, v ...interface{}) {
	if level < ERROR {
		return
	}
	output(ERROR, fmt.Sprintf(format, v...))
}

func Warning(msg string) {
	if level < WARNING {
		return
	}
	output(WARNING, msg)
}

func Warningf(format string, v ...interface{}) {
	if level < WARNING {
		return
	}
	output(WARNING, fmt.Sprintf(format, v...))
}

func Notice(msg string) {
	if level < NOTICE {
		return
	}
	output(NOTICE, msg)
}

func Noticef(format string, v ...interface{}) {
	if level < NOTICE {
		return
	}
	output(NOTICE, fmt.Sprintf(format, v...))
}

func Info(msg string) {
	if level < INFO {
		return
	}
	output(INFO, msg)
}

func Infof(format string, v ...interface{}) {
	if level < INFO {
		return
	}
	output(INFO, fmt.Sprintf(format, v...))
}

func Debug(msg string) {
	if level < DEBUG {
		return
	}
	output(DEBUG, msg)
}

func Debugf(format string, v ...interface{}) {
	if level < DEBUG {
		return
	}
	output(DEBUG, fmt.Sprintf(format, v...))
}
