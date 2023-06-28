package logger

// Testing:
// go test -coverprofile="coverage.out" -v ./...
// go tool cover -html="coverage.out"

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	LoggerSuite struct {
		suite.Suite
	}
)

func (s LoggerSuite) TestTracking() {
	t := Tracking(1)
	s.Equal("logger_test.go", t.File)
	s.NotEqual(0, t.Line)
	s.Equal("logger.LoggerSuite.TestTracking", t.Function)

	func() {
		t := Tracking(1)
		s.Equal("logger.LoggerSuite.TestTracking.func1", t.Function)
	}()

	func() {
		t := Tracking(2)
		s.Equal("logger.LoggerSuite.TestTracking", t.Function)
	}()
}

func (s LoggerSuite) TestDebug() {
	l := Logger{Level: DebugLevel}
	output := captureOutput(func() {
		l.Debug("TestDebug")
	})
	s.Contains(output, "[DEBUG]")
	s.Contains(output, "File: logger_test.go")
	s.Contains(output, "Function: logger.LoggerSuite.TestDebug.func1")
	s.Contains(output, "Message: [TestDebug]")
}

func (s LoggerSuite) TestInfo() {
	l := Logger{Level: InfoLevel}
	output := captureOutput(func() {
		l.Info("TestInfo")
	})
	s.Contains(output, "[INFO]")
	s.Contains(output, "File: logger_test.go")
	s.Contains(output, "Function: logger.LoggerSuite.TestInfo.func1")
	s.Contains(output, "Message: [TestInfo]")
}

func (s LoggerSuite) TestWarning() {
	l := Logger{Level: WarningLevel}
	output := captureOutput(func() {
		l.Warning("TestWarning")
	})
	s.Contains(output, "[WARNING]")
	s.Contains(output, "File: logger_test.go")
	s.Contains(output, "Function: logger.LoggerSuite.TestWarning.func1")
	s.Contains(output, "Message: [TestWarning]")
}

func (s LoggerSuite) TestError() {
	l := Logger{Level: ErrorLevel}
	output := captureOutput(func() {
		l.Error(fmt.Errorf("%s", "[TestError]"))
	})
	s.Contains(output, "[ERROR]")
	s.Contains(output, "File: logger_test.go")
	s.Contains(output, "Function: logger.LoggerSuite.TestError.func1")
	s.Contains(output, "Message: [TestError]")
}

func (s LoggerSuite) TestFatal() {
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
			l := Logger{Level: FatalLevel}
			output := captureOutput(func() {
				l.Fatal(fmt.Errorf("%s", "[TestFatal]"))
			})
			s.Contains(output, "[FATAL]")
			s.Contains(output, "File: logger_test.go")
			s.Contains(output, "Function: logger.LoggerSuite.TestFatal.func1")
			s.Contains(output, "Message: [TestFatal]")
		},
		"Fatal function was not called")
}

func (s LoggerSuite) TestGetFuncName() {
	fn := getFuncName(0)
	s.Equal("logger.LoggerSuite.TestGetFuncName", fn)

	unknown := getFuncName(100)
	s.Equal("unknown", unknown)
}

func (s LoggerSuite) TestFetchNameFromPath() {
	path := "/path/fo/testFuncName"

	funcName := fetchNameFromPath(path)
	s.Equal("testFuncName", funcName)

	path = "testFuncName"

	funcName = fetchNameFromPath(path)
	s.Equal("testFuncName", funcName)
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

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}
