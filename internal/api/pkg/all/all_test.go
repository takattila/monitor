package all

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/run"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiAllSuite struct {
		suite.Suite
	}
)

func (a ApiAllSuite) TestGetJSON() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", false)
	s.Data.Set("Services", false)
	s.Data.Set("TopProcesses", false)
	s.Data.Set("NetworkTraffic", false)
	s.Data.Set("Storage", false)

	cpu.Cfg, memory.Cfg, model.Cfg, network.Cfg, processes.Cfg, run.Cfg, services.Cfg, storage.Cfg = s, s, s, s, s, s, s, s

	r := GetRawJSONs()
	JSON := r.GetJSON()
	a.Contains(JSON, "model_name")
	a.Contains(JSON, "processor_info")
	a.Contains(JSON, "storage_info")
	a.Contains(JSON, "process_info")
	a.Contains(JSON, "services_info")
	a.Contains(JSON, "network_info")
	a.Contains(JSON, "uptime_info")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)

}

func (a ApiAllSuite) TestGetJSONMergeJSONError() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", false)
	s.Data.Set("Services", false)
	s.Data.Set("TopProcesses", false)
	s.Data.Set("NetworkTraffic", false)
	s.Data.Set("Storage", false)

	cpu.Cfg, memory.Cfg, model.Cfg, network.Cfg, processes.Cfg, run.Cfg, services.Cfg, storage.Cfg = s, s, s, s, s, s, s, s

	oldGetRawJSONs := GetRawJSONs
	GetRawJSONs := func() *AllJSONs {
		RawJSONs := []json.RawMessage{
			json.RawMessage(model.GetJSON()),
			json.RawMessage(cpu.GetJSON()),
			json.RawMessage(memory.GetJSON()),
			json.RawMessage(`{"storage_info": bad}`),
			json.RawMessage(processes.GetJSON()),
			json.RawMessage(services.GetJSON()),
			json.RawMessage(network.GetJSON()),
		}
		return &AllJSONs{RawJSONs: RawJSONs}
	}
	r := GetRawJSONs()
	a.NotContains("storage_info", r.GetJSON())

	GetRawJSONs = oldGetRawJSONs
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiAllSuite(t *testing.T) {
	suite.Run(t, new(ApiAllSuite))
}
