// plog is a simple "print" logger that writes to stdout.
package plog

import (
	"fmt"
	"strconv"
	"time"
)

const (
	levelDebug   = "DBG"
	levelInfo    = "INF"
	levelWarning = "WRN"
	levelError   = "ERR"
	levelFatal   = "FTL"

	reset = "\033[0m"

	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97

	timeFormat = "02-Jan-06 15:04:05"
)

func writeLog(level string, color int, format string, args ...any) {
	fmt.Printf("\033[%sm[%s] [%s]: %s%s\n",
		strconv.Itoa(color),
		time.Now().UTC().Format(timeFormat),
		level,
		fmt.Sprintf(format, args...),
		reset)
}

func Debug(msg string) {
	writeLog(levelDebug, lightGray, "%s", msg)
}

func Debugf(format string, args ...any) {
	writeLog(levelDebug, lightGray, format, args...)
}

func Info(msg string) {
	writeLog(levelInfo, green, "%s", msg)
}

func Infof(format string, args ...any) {
	writeLog(levelInfo, green, format, args...)
}

func Warning(msg string) {
	writeLog(levelWarning, yellow, "%s", msg)
}

func Warningf(format string, args ...any) {
	writeLog(levelWarning, yellow, format, args...)
}

func Error(msg string) {
	writeLog(levelError, red, "%s", msg)
}

func Errorf(format string, args ...any) {
	writeLog(levelError, red, format, args...)
}

func Fatal(msg string) {
	writeLog(levelFatal, lightRed, "%s", msg)
}

func Fatalf(format string, args ...any) {
	writeLog(levelFatal, lightRed, format, args...)
}
