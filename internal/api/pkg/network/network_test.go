package network

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
	ApiNetworkSuite struct {
		suite.Suite
	}
)

func (a ApiNetworkSuite) TestGetJSON() {
	Sleep = 10 * time.Millisecond

	s := getConfig("api", "linux")
	s.Data.Set("NetworkTraffic", true)
	Cfg = s

	go Stats()
	time.Sleep(100 * time.Millisecond)

	JSON := GetJSON()
	a.Contains(JSON, "network_info")

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

func TestApiNetworkSuite(t *testing.T) {
	suite.Run(t, new(ApiNetworkSuite))
}
