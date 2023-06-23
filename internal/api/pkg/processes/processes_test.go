package processes

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiProcessesSuite struct {
		suite.Suite
	}
)

func (a ApiProcessesSuite) TestGetJSONByConfigCommand() {
	s := getConfig("api", "linux")

	for _, toggle := range []bool{true, false} {
		s.Data.Set("TopProcesses", toggle)
		Cfg = s

		JSON := GetJSON()
		a.Contains(JSON, "process_info")

		d := make(map[string]interface{})
		err := json.Unmarshal([]byte(JSON), &d)
		a.Equal(err, nil)
	}
}

func (a ApiProcessesSuite) TestGetJSONWithoutConfigCommand() {
	s := getConfig("api", "linux")

	s.Data.Set("TopProcesses", true)
	s.Data.Set("on_runtime.commands.processes", []string{})
	Cfg = s

	JSON := GetJSON()
	a.Contains(JSON, "process_info")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)
}

func (a ApiProcessesSuite) TestGetFunctions() {
	result := getPid("")
	a.Equal("0", result)
	result = getPid(" ")
	a.Equal("0", result)
	result = getPid("1")
	a.Equal("1", result)

	result = getUser("1")
	a.Equal("unknown", result)
	result = getUser("1 user")
	a.Equal("user", result)

	result = getMem("1 user")
	a.Equal("0.0%", result)
	result = getMem("1 user 25.0%")
	a.Equal("25.0%", result)

	result = getCpu("1 user 25.0%")
	a.Equal("0.0%", result)
	result = getCpu("1 user 25.0% 45.0%")
	a.Equal("45.0%", result)

	result = getCmd("1 user 25.0% 45.0%")
	a.Equal("unknown", result)
	result = getCmd("1 user 25.0% 45.0% command")
	a.Equal("command", result)
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiProcessesSuite(t *testing.T) {
	suite.Run(t, new(ApiProcessesSuite))
}
