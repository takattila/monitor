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
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/takattila/monitor/pkg/logger"
)

// GetConfigPathCmd contains the command that fetches the config path by hostname.
var GetConfigPathCmd = []string{"bash", "-c", "hostnamectl | grep Operating | awk -F: '{print $2}' | xargs"}

// DynamicSizeSI returns the value of the given number with unit dynamically.
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

// DynamicSizeSI returns the size only of the given number with unit dynamically.
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

// DynamicSizeSI returns the unit only of the given number dynamically.
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

// DynamicSizeIEC returns the value of the given number with unit dynamically.
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

// DynamicSizeIECSize returns the size only of the given number with unit dynamically.
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

// DynamicSizeIECUnit returns the unit only of the given number dynamically.
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

// GetPercent calculates what percentage of the number 'a' is of the number 'b'
func GetPercent(a uint64, b uint64) float64 {
	decimalPlaces := 1
	result := (float64(a) / float64(b) * 100)
	ratio := math.Pow(10, float64(decimalPlaces))
	return math.Round(result*ratio) / ratio
}

// Cli issues a command passed as a string slice.
func Cli(command []string) string {
	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	ErrorIfErr(err)
	return out.String()
}

// GetProgramDir returns with the directory of the current program.
func GetProgramDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	base := filepath.Base(dir)
	ErrorIfErr(err)
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

// GetString fetches a alphanumeric characters only.
func GetString(str string) string {
	re := regexp.MustCompile(`[^a-zA-Z]*`)
	return re.ReplaceAllString(str, "")
}

// GetNum fetches numbers only from the string.
func GetNum(str string) uint64 {
	re := regexp.MustCompile(`[^0-9]*`)
	num, err := strconv.ParseUint(re.ReplaceAllString(str, ""), 10, 64)
	ErrorIfErr(err)
	return num
}

// TextToBytes converts a test to bytes, for example: TextToBytes("1kB") = 1024
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

// ReplaceStringInSlice replaces a string in a slice.
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

// SliceContains checks wheter a slice contains a string or not.
func SliceContains(slice []string, contains string) bool {
	for _, elem := range slice {
		if elem == contains {
			return true
		}
	}
	return false
}

// GetConfigPath gives back the path of the service configuration file.
// It fetches the kind of configuration file from the name of the OS.
func GetConfigPath(service string) string {
	check := Cli(GetConfigPathCmd)
	if strings.Contains(strings.ToLower(check), "raspbian") {
		return filepath.Join(GetProgramDir(), "/configs/"+service+".raspbian.yaml")
	}
	return filepath.Join(GetProgramDir(), "/configs/"+service+".linux.yaml")
}

// ErrorIfErr prints error message if the 'err' argument is not nil.
func ErrorIfErr(err error) {
	if err != nil {
		track := logger.Tracking(2)
		log.Println(color.HiRedString("[ERROR]"), color.HiRedString("File:"), track.File, color.HiRedString("Function:"), track.Function, color.HiRedString("Line:"), track.Line, color.HiRedString("Message:"), err)
	}
}
