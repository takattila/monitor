package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

// LogLevel sets the tier of the logging.
type LogLevel int

const (
	// DebugLevel represents the Debug function.
	DebugLevel LogLevel = iota
	// InfoLevel represents the Info function.
	InfoLevel
	// WarningLevel represents the Warning function.
	WarningLevel
	// ErrorLevel represents the Error function.
	ErrorLevel
	// FatalLevel represents the Fatal function.
	FatalLevel
)

// Logger initializes the tier of the logger functions.
type Logger struct {
	Level LogLevel
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

// Debug writes Debug to stdOut.
func (l Logger) Debug(args ...interface{}) {
	if l.Level <= DebugLevel {
		track := getTrackingInfo(1)
		log.Println("[DEBUG] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", args)
	}
}

// Info writes Info to stdOut.
func (l Logger) Info(args ...interface{}) {
	if l.Level <= InfoLevel {
		track := getTrackingInfo(1)
		log.Println("[INFO] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", args)
	}
}

// Info writes Warning to stdOut.
func (l Logger) Warning(args ...interface{}) {
	if l.Level <= WarningLevel {
		track := getTrackingInfo(1)
		log.Println("[WARNING] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", args)
	}
}

// Info writes Error to stdOut.
func (l Logger) Error(err error) {
	if l.Level <= ErrorLevel {
		if err != nil {
			track := getTrackingInfo(1)
			log.Println("[ERROR] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", err)
		}
	}
}

// Fatal writes Error to stdOut and exit with exit code 1, if err doesn't nil.
func (l Logger) Fatal(err error) {
	if l.Level <= FatalLevel {
		if err != nil {
			track := getTrackingInfo(1)
			log.Println("[FATAL] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", err)
			os.Exit(1)
		}
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
