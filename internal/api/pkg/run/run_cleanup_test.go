package run

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiRunCleanupSuite struct {
		suite.Suite
	}
)

func (a ApiRunCleanupSuite) TestCleanup() {
	oldCmdFolder := CmdFolder
	CmdFolder = gitRootPath + "/cmd/"
	defer func() {
		Cleanup()
		CmdFolder = oldCmdFolder
	}()

	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	err := Exec("get_storages")
	a.Equal(nil, err)

	content := StdOut("get_storages")
	a.NotEqual("", content)

	Cleanup()

	check, err := checkFiletypesInDir("stdout")
	a.Equal(nil, err)
	a.Equal(false, check)

	check, err = checkFiletypesInDir("finish")
	a.Equal(nil, err)
	a.Equal(false, check)

}

func checkFiletypesInDir(ext string) (bool, error) {
	ext = "." + ext

	d, err := os.Open(CmdFolder)
	if err != nil {
		return false, err
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ext {
				return true, nil
			}
		}
	}
	return false, nil
}

func TestApiRunCleanupSuite(t *testing.T) {
	suite.Run(t, new(ApiRunCleanupSuite))
}
