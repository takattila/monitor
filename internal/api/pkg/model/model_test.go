package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiModelSuite struct {
		suite.Suite
	}
)

func (a ApiModelSuite) TestGetJSON() {
	s := getConfig("api", "linux")
	Cfg = s

	JSON := GetJSON()
	a.Contains(JSON, "model_name")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)
}

func (a ApiModelSuite) TestGetModelNameFromOS() {
	s := getConfig("api", "linux")
	s.Data.Set("on_runtime.commands.model_name", "")
	Cfg = s

	m := Model{}
	m.getModelName()
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiModelSuite(t *testing.T) {
	suite.Run(t, new(ApiModelSuite))
}
