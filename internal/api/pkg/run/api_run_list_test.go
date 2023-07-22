package run

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
)

type (
	ApiRunListSuite struct {
		suite.Suite
	}
)

func (a ApiRunListSuite) TestGetJSON() {
	Cfg = getConfig("api", "linux")
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	JSON := GetJSON()
	a.Contains(JSON, "run_list")
}

func (a ApiRunListSuite) TestGetJSONWithoutRuns() {
	Cfg = getConfig("api", "linux")
	Cfg.Data.Set("on_runtime.run", []string{})
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	JSON := GetJSON()
	a.Contains(JSON, "run_list")
}

func TestApiRunListSuite(t *testing.T) {
	suite.Run(t, new(ApiRunListSuite))
}
