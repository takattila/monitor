package skins

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	ApiSkinSuite struct {
		suite.Suite
	}
)

var (
	gitRootPath = strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
)

func (a ApiSkinSuite) TestGetJSON() {
	s := getConfig("api", "linux")
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	oldSkinsPath := SkinsPath
	SkinsPath = gitRootPath + "/web/css"
	defer func() { SkinsPath = oldSkinsPath }()

	JSON := GetJSON()
	a.Contains(JSON, "skins")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)
}

func getConfig(service, system string) *settings.Settings {
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiSkinSuite(t *testing.T) {
	suite.Run(t, new(ApiSkinSuite))
}
