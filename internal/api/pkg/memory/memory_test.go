package memory

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	ApiMemorySuite struct {
		suite.Suite
	}
)

func (a ApiMemorySuite) TestGetJSON() {
	s := getConfig("api", "linux")

	for _, toggle := range []bool{true, false} {
		s.Data.Set("Memory", toggle)
		Cfg = s
		L = logger.New(logger.NoneLevel, logger.ColorOff)

		JSON := GetJSON()
		a.Contains(JSON, "memory_info")
		a.Contains(JSON, "total")
		a.Contains(JSON, "used")
		a.Contains(JSON, "free")
		a.Contains(JSON, "cached")
		a.Contains(JSON, "available")
		a.Contains(JSON, "swap")
		a.Contains(JSON, "video")

		d := make(map[string]interface{})
		err := json.Unmarshal([]byte(JSON), &d)
		a.Equal(err, nil)
	}
}

func (a ApiMemorySuite) TestGetPhysicalMemoryFromConfiguration() {
	s := getConfig("api", "linux")

	s.Data.Set("Memory", true)
	s.Data.Set("on_runtime.physical_memory", "1GB")
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	vm, err := mem.VirtualMemory()
	a.Equal(nil, err)
	a.Equal(getPhysicalMemory(vm), uint64(1024*1024*1024))
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestApiMemorySuite(t *testing.T) {
	suite.Run(t, new(ApiMemorySuite))
}
