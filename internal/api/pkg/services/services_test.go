package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiServicesSuite struct {
		suite.Suite
	}
)

func (a ApiServicesSuite) TestGetJSON() {
	Sleep = 10 * time.Millisecond

	s := getConfig("api", "linux")
	s.Data.Set("Services", true)
	Cfg = s

	go Watcher()
	time.Sleep(100 * time.Millisecond)

	JSON := GetJSON()
	a.Contains(JSON, "services_info")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiServicesSuite(t *testing.T) {
	suite.Run(t, new(ApiServicesSuite))
}
