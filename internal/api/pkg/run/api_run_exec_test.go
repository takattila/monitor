package run

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiRunExecSuite struct {
		suite.Suite
	}
)

func (a ApiRunExecSuite) TestExec() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	Cleanup()

	err := Exec("get_storages")
	a.Equal(nil, err)
}

func (a ApiRunExecSuite) TestRunOsCreateError() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cleanup()

	oldOsCreate := osCreate
	osCreate = func(stdout string) (*os.File, error) {
		return nil, fmt.Errorf("osCreate %s", "error")
	}
	defer func() { osCreate = oldOsCreate }()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	err := Exec("get_storages")
	a.Equal(fmt.Errorf("osCreate %s", "error"), err)
}

func (a ApiRunExecSuite) TestRunCmdStdoutPipeError() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cleanup()

	oldCmdStdoutPipe := cmdStdoutPipe
	cmdStdoutPipe = func(cmd *exec.Cmd) (io.ReadCloser, error) {
		return nil, fmt.Errorf("cmdStdoutPipe %s", "error")
	}
	defer func() { cmdStdoutPipe = oldCmdStdoutPipe }()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	err := Exec("get_storages")
	a.Equal(fmt.Errorf("cmdStdoutPipe %s", "error"), err)
}

func (a ApiRunExecSuite) TestRunCmdStartError() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cleanup()

	oldCmdStart := cmdStart
	cmdStart = func(cmd *exec.Cmd) error {
		return fmt.Errorf("cmdStart %s", "error")
	}
	defer func() { cmdStart = oldCmdStart }()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	err := Exec("get_storages")
	a.Equal(fmt.Errorf("cmdStart %s", "error"), err)
}

func (a ApiRunExecSuite) TestRunStdoutExistsAndFinishNotExistsError() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	Cleanup()

	err := createFile(CmdFolder + "get_storages.stdout")
	a.Equal(nil, err)

	err = Exec("get_storages")
	a.Contains(fmt.Sprint(err), "is running  already")
}

func (a ApiRunExecSuite) TestRunStdoutExistsAndFinishExistsError() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	Cleanup()

	err := createFile(CmdFolder + "get_storages.stdout")
	a.Equal(nil, err)

	err = createFile(CmdFolder + "get_storages.finish")
	a.Equal(nil, err)

	err = Exec("get_storages")
	a.Equal(nil, err)
}

func createFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func TestApiRunExecSuite(t *testing.T) {
	suite.Run(t, new(ApiRunExecSuite))
}
