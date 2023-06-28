package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// LogLevel sets the tier of the logging.
type LogLevel int

const (
	// NoneLevel suppress all logger functions.
	NoneLevel LogLevel = iota
	// FatalLevel represents the Fatal function.
	FatalLevel
	// ErrorLevel represents the Error function.
	ErrorLevel
	// WarningLevel represents the Warning function.
	WarningLevel
	// InfoLevel represents the Info function.
	InfoLevel
	// DebugLevel represents the Debug function.
	DebugLevel
)

// Color turns color mode on or off.
type Color int

const (
	// ColorOff turns off color mode.
	ColorOff Color = iota
	// ColorOff turns on color mode.
	ColorOn
)

// Logger initializes the tier of the logger functions.
type Logger struct {
	Level    LogLevel
	Colorize Color
}

// TrackInfo holds debug information about the caller's:
//   - File name
//   - Line number
//   - Function name
type TrackInfo struct {
	File     string
	Line     string
	Function string
}

// New provides a Logger struct.
func New(level LogLevel, color Color) Logger {
	return Logger{
		Level:    level,
		Colorize: color,
	}
}

// Debug writes Debug to stdOut.
func (l Logger) Debug(args ...interface{}) {
	if l.Level >= DebugLevel {
		track := getTrackingInfo(1)
		l.print(color.HiGreenString, "debug", track.File, track.Function, track.Line, args...)
	}
}

// Info writes Info to stdOut.
func (l Logger) Info(args ...interface{}) {
	if l.Level >= InfoLevel {
		track := getTrackingInfo(1)
		l.print(color.HiBlueString, "info", track.File, track.Function, track.Line, args...)
	}
}

// Warning writes Warning to stdOut.
func (l Logger) Warning(args ...interface{}) {
	if l.Level >= WarningLevel {
		track := getTrackingInfo(1)
		l.print(color.HiYellowString, "warning", track.File, track.Function, track.Line, args...)
	}
}

// Error writes Error to stdOut.
func (l Logger) Error(err error) {
	if l.Level >= ErrorLevel && err != nil {
		track := getTrackingInfo(1)
		l.print(color.HiRedString, "error", track.File, track.Function, track.Line, err)
	}
}

// Fatal writes Error to stdOut and exit with exit code 1, if err doesn't nil.
func (l Logger) Fatal(err error) {
	if l.Level >= FatalLevel && err != nil {
		track := getTrackingInfo(1)
		l.print(color.HiBlackString, "fatal", track.File, track.Function, track.Line, err)
		os.Exit(1)
	}
}

// Tracking provides debug information about function invocations
// on the calling goroutine's stack:
//   - File (where Tracking was called)
//   - Line (where function was called)
//   - Function (name of the function)
func Tracking(depth int) TrackInfo {
	return getTrackingInfo(depth)
}

// getTrackingInfo provides debug information about function invocations
// on the calling goroutine's stack:
//   - File (where Tracking was called)
//   - Line (where function was called)
//   - Function (name of the function)
func getTrackingInfo(depth int) TrackInfo {
	_, fileName, line, _ := runtime.Caller(depth + 1)
	return TrackInfo{
		File:     fetchNameFromPath(fileName),
		Line:     fmt.Sprintf("%d", line),
		Function: getFuncName(depth + 1),
	}
}

// getFuncName returns with the caller's function name
func getFuncName(depth int) string {
	pc, _, _, _ := runtime.Caller(depth + 1)
	me := runtime.FuncForPC(pc)
	if me == nil {
		return "unknown"
	}
	return fetchNameFromPath(me.Name())
}

// fetchNameFromPath extracts the name of a function from a path.
func fetchNameFromPath(fileName string) string {
	for i := len(fileName) - 1; i > 0; i-- {
		if fileName[i] == '/' {
			return fileName[i+1:]
		}
	}
	return fileName
}

// print writes logging messages into stdOut and also decides whether the color functionality should be turned on or off.
func (l Logger) print(c func(format string, a ...interface{}) string, level, file, function, line string, args ...interface{}) {
	if l.Colorize == ColorOn {
		log.Println(c("["+strings.ToUpper(level)+"]"), c("File:"), file, c("Function:"), function, c("Line:"), line, c("Message:"), args)
	} else {
		log.Println("["+strings.ToUpper(level)+"]", "File:", file, "Function:", function, "Line:", line, "Message:", args)
	}
}
