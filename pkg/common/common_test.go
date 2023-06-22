package common

// Testing:
// go test -coverprofile="coverage.out" -v ./...
// go tool cover -html="coverage.out"

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	Suite struct {
		suite.Suite
	}
)

func (s Suite) TestTracking() {
	t := Tracking(1)
	s.Equal("common_test.go", t.File)
	s.NotEqual(0, t.Line)
	s.Equal("common.Suite.TestTracking", t.Function)

	func() {
		t := Tracking(1)
		s.Equal("common.Suite.TestTracking.func1", t.Function)
	}()

	func() {
		t := Tracking(2)
		s.Equal("common.Suite.TestTracking", t.Function)
	}()
}

func (s Suite) TestInfo() {
	output := captureOutput(func() {
		Info("TestInfo")
	})
	s.Contains(output, "[INFO]")
	s.Contains(output, "File: common_test.go")
	s.Contains(output, "Function: common.Suite.TestInfo.func1")
	s.Contains(output, "Message: [TestInfo]")
}

func (s Suite) TestWarning() {
	output := captureOutput(func() {
		Warning("TestWarning")
	})
	s.Contains(output, "[WARNING]")
	s.Contains(output, "File: common_test.go")
	s.Contains(output, "Function: common.Suite.TestWarning.func1")
	s.Contains(output, "Message: [TestWarning]")
}

func (s Suite) TestError() {
	output := captureOutput(func() {
		Error(fmt.Errorf("%s", "[TestError]"))
	})
	s.Contains(output, "[ERROR]")
	s.Contains(output, "File: common_test.go")
	s.Contains(output, "Function: common.Suite.TestError.func1")
	s.Contains(output, "Message: [TestError]")
}

func (s Suite) TestFatal() {
	ExpectedPanicText := "Fatal function called"

	panicFunc := func(int) {
		panic(ExpectedPanicText)
	}

	patch := monkey.Patch(os.Exit, panicFunc)
	defer patch.Unpatch()

	assert.PanicsWithValue(
		s.T(),
		ExpectedPanicText,
		func() {
			output := captureOutput(func() {
				Fatal(fmt.Errorf("%s", "[TestFatal]"))
			})
			s.Contains(output, "[FATAL]")
			s.Contains(output, "File: common_test.go")
			s.Contains(output, "Function: common.Suite.TestFatal.func1")
			s.Contains(output, "Message: [TestFatal]")
		},
		"Fatal function was not called")
}

func (s Suite) TestGetFuncName() {
	fn := getFuncName(0)
	s.Equal("common.Suite.TestGetFuncName", fn)

	unknown := getFuncName(100)
	s.Equal("unknown", unknown)
}

func (s Suite) TestFetchNameFromPath() {
	path := "/path/fo/testFuncName"

	funcName := fetchNameFromPath(path)
	s.Equal("testFuncName", funcName)

	path = "testFuncName"

	funcName = fetchNameFromPath(path)
	s.Equal("testFuncName", funcName)
}

func (s Suite) TestDynamicSizeSI() {
	size := DynamicSizeSI(uint64(1))
	s.Equal("1 B", size)

	size = DynamicSizeSI(uint64(1000))
	s.Equal("1.0 kB", size)

	size = DynamicSizeSI(uint64(1000000))
	s.Equal("1.0 MB", size)
}

func (s Suite) TestDynamicSizeSISize() {
	size := DynamicSizeSISize(uint64(1))
	s.Equal(float64(1), size)

	size = DynamicSizeSISize(uint64(1000))
	s.Equal(float64(1), size)

	size = DynamicSizeSISize(uint64(1000000))
	s.Equal(float64(1), size)
}

func (s Suite) TestDynamicSizeSIUnit() {
	size := DynamicSizeSIUnit(uint64(1))
	s.Equal("B", size)

	size = DynamicSizeSIUnit(uint64(1000))
	s.Equal("kB", size)

	size = DynamicSizeSIUnit(uint64(1000000))
	s.Equal("MB", size)
}

func (s Suite) TestDynamicSizeIEC() {
	size := DynamicSizeIEC(uint64(1))
	s.Equal("1 B", size)

	size = DynamicSizeIEC(uint64(1024))
	s.Equal("1.0 kB", size)

	size = DynamicSizeIEC(uint64(1048576))
	s.Equal("1.0 MB", size)
}

func (s Suite) TestDynamicSizeIECSize() {
	size := DynamicSizeIECSize(uint64(1))
	s.Equal(float64(1), size)

	size = DynamicSizeIECSize(uint64(1024))
	s.Equal(float64(1), size)

	size = DynamicSizeIECSize(uint64(1048576))
	s.Equal(float64(1), size)
}

func (s Suite) TestDynamicSizeIECUnit() {
	size := DynamicSizeIECUnit(uint64(1))
	s.Equal("B", size)

	size = DynamicSizeIECUnit(uint64(1024))
	s.Equal("kB", size)

	size = DynamicSizeIECUnit(uint64(1048576))
	s.Equal("MB", size)
}

func (s Suite) TestGetPercent() {
	percent := GetPercent(uint64(50), uint64(100))
	s.Equal(float64(50), percent)
}

func (s Suite) TestCli() {
	cmd := []string{""}
	if runtime.GOOS == "windows" {
		cmd = []string{"powershell", "-NoProfile", "-c", "echo 'Hello World'"}
	} else {
		cmd = []string{"bash", "-c", "echo 'Hello World'"}

	}
	output := Cli(cmd)
	s.Contains(output, "Hello World")
}

func (s Suite) TestGetProgramDir() {
	dir := GetProgramDir()
	s.NotEmpty(dir)
}

func (s Suite) TestFileExists() {
	s.Equal(true, FileExists("common_test.go"))

	s.Equal(false, FileExists("random_file.go"))

	s.Equal(false, FileExists("../_codegen"))
}

func (s Suite) TestGetString() {
	str := GetString("56465456HELLO757657")
	s.Equal("HELLO", str)
}

func (s Suite) TestGetNum() {
	num := GetNum("1234HELLO5678")
	s.Equal(uint64(12345678), num)
}

func (s Suite) TestTextToBytes() {
	size := 2

	text := TextToBytes(fmt.Sprintf("%dPB", size))
	num := size * 1024 * 1024 * 1024 * 1024 * 1024
	s.Equal(uint64(num), text)

	text = TextToBytes(fmt.Sprintf("%dTB", size))
	num = size * 1024 * 1024 * 1024 * 1024
	s.Equal(uint64(num), text)

	text = TextToBytes(fmt.Sprintf("%dGB", size))
	num = size * 1024 * 1024 * 1024
	s.Equal(uint64(num), text)

	text = TextToBytes(fmt.Sprintf("%dMB", size))
	num = size * 1024 * 1024
	s.Equal(uint64(num), text)

	text = TextToBytes(fmt.Sprintf("%dKB", size))
	num = size * 1024
	s.Equal(uint64(num), text)

	text = TextToBytes(fmt.Sprintf("%dB", size))
	num = size
	s.Equal(uint64(num), text)

	text = TextToBytes("0")
	num = 0
	s.Equal(uint64(num), text)
}

func (s Suite) TestReplaceStringInSlice() {
	slice := ReplaceStringInSlice([]string{"a", "b", "c"}, "c", "new")
	s.Equal([]string{"a", "b", "new"}, slice)
}

func (s Suite) TestSliceContains() {
	result := SliceContains([]string{"a", "b", "c"}, "c")
	s.Equal(true, result)

	result = SliceContains([]string{"a", "b", "c"}, "d")
	s.Equal(false, result)
}

func (s Suite) TestGetConfigPath() {
	OldGetConfigPathCmd := GetConfigPathCmd
	NewGetConfigPathCmd := OldGetConfigPathCmd

	if runtime.GOOS == "windows" {
		NewGetConfigPathCmd = []string{"powershell", "-NoProfile", "-c"}
	} else {
		NewGetConfigPathCmd = []string{"bash", "-c"}

	}

	GetConfigPathCmd = append(NewGetConfigPathCmd, "echo raspbian")
	result := GetConfigPath("web")
	s.Contains(result, "web.raspbian.yaml")

	GetConfigPathCmd = append(NewGetConfigPathCmd, "echo linux")
	result = GetConfigPath("web")
	s.Contains(result, "web.linux.yaml")

	GetConfigPathCmd = OldGetConfigPathCmd
}

func captureOutput(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
		log.SetOutput(os.Stderr)
	}()
	os.Stdout = writer
	os.Stderr = writer
	log.SetOutput(writer)
	out := make(chan string)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, reader)
		out <- buf.String()
	}()
	wg.Wait()
	f()
	writer.Close()
	return <-out
}

func TestCommonSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}
