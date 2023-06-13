package common

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// -- logger ----------------------------------------------------------

// TrackInfo holds debug information about the caller's:
//   - File name
//   - Line number
//   - Function name
type TrackInfo struct {
	File     string
	Line     string
	Function string
}

func Info(args ...interface{}) {
	track := getTrackingInfo(1)
	log.Println("[INFO] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", args)
}

func Warning(args ...interface{}) {
	track := getTrackingInfo(1)
	log.Println("[WARNING] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", args)
}

func Error(err error) {
	if err != nil {
		track := getTrackingInfo(1)
		log.Println("[ERROR] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", err)
	}
}

func Fatal(err error) {
	if err != nil {
		track := getTrackingInfo(1)
		log.Println("[FATAL] File:", track.File, "Function:", track.Function, "Line:", track.Line, "Message:", err)
		os.Exit(1)
	}
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

// -- logger ----------------------------------------------------------

func DynamicSizeSI(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	decimalPlaces := 1
	ratio := math.Pow(10, float64(decimalPlaces))
	result := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %cB",
		math.Round(result*ratio)/ratio, "kMGTPE"[exp])
}

func DynamicSizeSISize(b uint64) float64 {
	const unit = 1000
	if b < unit {
		return float64(b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	decimalPlaces := 1
	ratio := math.Pow(10, float64(decimalPlaces))
	result := float64(b) / float64(div)
	return math.Round(result*ratio) / ratio
}

func DynamicSizeSIUnit(b uint64) string {
	const unit = 1000
	if b < unit {
		return "B"
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%cB", "kMGTPE"[exp])
}

func DynamicSizeIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	decimalPlaces := 1
	ratio := math.Pow(10, float64(decimalPlaces))
	result := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %cB",
		math.Round(result*ratio)/ratio, "kMGTPE"[exp])
}

func DynamicSizeIECSize(b uint64) float64 {
	const unit = 1024
	if b < unit {
		return float64(b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	decimalPlaces := 1
	ratio := math.Pow(10, float64(decimalPlaces))
	result := float64(b) / float64(div)
	return math.Round(result*ratio) / ratio
}

func DynamicSizeIECUnit(b uint64) string {
	const unit = 1024
	if b < unit {
		return "B"
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%cB", "kMGTPE"[exp])
}

func GetPercent(a uint64, b uint64) float64 {
	decimalPlaces := 1
	result := (float64(a) / float64(b) * 100)
	ratio := math.Pow(10, float64(decimalPlaces))
	return math.Round(result*ratio) / ratio
}

func Cli(command []string) string {
	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	Error(err)
	return out.String()
}

func GetProgramDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	base := filepath.Base(dir)
	Error(err)
	return strings.Replace(dir, string(os.PathSeparator)+base, "", 1)
}

// FileExists checks if a file exists and is not a directory.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info == nil {
		return false
	}
	return !info.IsDir()
}

func GetString(str string) string {
	re := regexp.MustCompile(`[^a-zA-Z]*`)
	return re.ReplaceAllString(str, "")
}

func GetNum(str string) uint64 {
	re := regexp.MustCompile(`[^0-9]*`)
	num, err := strconv.ParseUint(re.ReplaceAllString(str, ""), 10, 64)
	Error(err)
	return num
}

func TextToBytes(text string) uint64 {
	unit := GetString(text)
	size := GetNum(text)
	switch strings.ToUpper(unit) {
	case "PB":
		return size * 1024 * 1024 * 1024 * 1024 * 1024
	case "TB":
		return size * 1024 * 1024 * 1024 * 1024
	case "GB":
		return size * 1024 * 1024 * 1024
	case "MB":
		return size * 1024 * 1024
	case "KB":
		return size * 1024
	case "B":
		return size
	default:
		return 0
	}
}

func ReplaceStringInSlice(s []string, old string, new string) []string {
	newSlice := make([]string, 0)
	for _, v := range s {
		if strings.Contains(v, old) {
			newSlice = append(newSlice, strings.ReplaceAll(v, old, new))
		} else {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func SliceContains(slice []string, contains string) bool {
	for _, elem := range slice {
		if elem == contains {
			return true
		}
	}
	return false
}

func GetConfigPath(service string) string {
	cmd := []string{"bash", "-c", "hostnamectl | grep Operating | awk -F: '{print $2}' | xargs"}
	check := Cli(cmd)
	if strings.Contains(strings.ToLower(check), "raspbian") {
		return filepath.Join(GetProgramDir(), "/configs/"+service+".raspbian.yaml")
	}
	return filepath.Join(GetProgramDir(), "/configs/"+service+".linux.yaml")
}
