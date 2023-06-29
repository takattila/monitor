package services

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	ApiServicesSuite struct {
		suite.Suite
	}
)

func (a ApiServicesSuite) TestGetJSON() {
	Sleep = 10 * time.Millisecond
	oldPetProcessesStatus := getProcessesStatus
	defer func() { getProcessesStatus = oldPetProcessesStatus }()

	getProcessesStatus = func(services []string) (output string) {
		return `service1 active enabled
		service2 inactive disabled
		service3 active 
		service4 `
	}

	s := getConfig("api", "linux")
	s.Data.Set("Services", true)
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	go Watcher()
	time.Sleep(100 * time.Millisecond)

	JSON := GetJSON()
	a.Contains(JSON, "services_info")
	a.Contains(JSON, "service1")
	a.Contains(JSON, "service2")
	a.Contains(JSON, "service3")

	type ServicesInfo struct {
		ServicesInfo struct {
			Service1 struct {
				IsActive  string `json:"is_active"`
				IsEnabled string `json:"is_enabled"`
			} `json:"service1"`
			Service2 struct {
				IsActive  string `json:"is_active"`
				IsEnabled string `json:"is_enabled"`
			} `json:"service2"`
			Service3 struct {
				IsActive  string `json:"is_active"`
				IsEnabled string `json:"is_enabled"`
			} `json:"service3"`
			Service4 struct {
				IsActive  string `json:"is_active"`
				IsEnabled string `json:"is_enabled"`
			} `json:"service4"`
		} `json:"services_info"`
	}

	d := ServicesInfo{}

	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)

	a.Equal("active", d.ServicesInfo.Service1.IsActive)
	a.Equal("enabled", d.ServicesInfo.Service1.IsEnabled)

	a.Equal("inactive", d.ServicesInfo.Service2.IsActive)
	a.Equal("disabled", d.ServicesInfo.Service2.IsEnabled)

	a.Equal("unknown", d.ServicesInfo.Service3.IsActive)
	a.Equal("unknown", d.ServicesInfo.Service3.IsEnabled)

	a.Equal("unknown", d.ServicesInfo.Service4.IsActive)
	a.Equal("unknown", d.ServicesInfo.Service4.IsEnabled)
}

func (a ApiServicesSuite) TestGetProcessesStatus() {
	output := getProcessesStatus([]string{
		"bad_service1",
		"bad_service2",
		"bad_service3",
	})

	a.Contains(output, "bad_service1 inactive")
	a.Contains(output, "bad_service2 inactive")
	a.Contains(output, "bad_service3 inactive")
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
