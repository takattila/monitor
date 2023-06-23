package storage

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiStorageSuite struct {
		suite.Suite
	}
)

func (a ApiStorageSuite) TestGetJSON() {
	for _, toggle := range []bool{true, false} {
		s := getConfig("api", "linux")
		s.Data.Set("Storage", toggle)
		Cfg = s

		JSON := GetJSON()
		a.Contains(JSON, "storage_info")

		d := make(map[string]interface{})
		err := json.Unmarshal([]byte(JSON), &d)
		a.Equal(err, nil)
	}
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiStorageSuite(t *testing.T) {
	suite.Run(t, new(ApiStorageSuite))
}
