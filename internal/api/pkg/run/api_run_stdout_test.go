package run

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiRunStdOutSuite struct {
		suite.Suite
	}
)

func (a ApiRunStdOutSuite) TestStdOutFinished() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cleanup()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	err := Exec("get_storages")
	a.Equal(nil, err)

	content := StdOut("get_storages")
	a.NotEqual("", content)
}

func (a ApiRunStdOutSuite) TestStdOutRunning() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cleanup()

	go func() {
		err := Exec("ping_10_localhost")
		a.Equal(nil, err)
	}()

	time.Sleep(600 * time.Millisecond)

	content := StdOut("ping_10_localhost")
	a.NotEqual("", content)
}

func TestApiRunStdOutSuite(t *testing.T) {
	suite.Run(t, new(ApiRunStdOutSuite))
}
