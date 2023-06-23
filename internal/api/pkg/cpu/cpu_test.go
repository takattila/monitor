package cpu

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	ApiAllSuite struct {
		suite.Suite
	}
)

func (a ApiAllSuite) TestGetJSONWithGetTempByUsage() {
	s := getConfig("api", "linux")
	Cfg = s

	JSON := GetJSON()
	a.Contains(JSON, "processor_info")
	a.Contains(JSON, "temp")
	a.Contains(JSON, "usage")
	a.Contains(JSON, "load")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)

}

func (a ApiAllSuite) TestGetJSONWithGetTemp() {
	s := getConfig("api", "linux")
	s.Data.Set("on_runtime.commands.cpu_temp", []string{"bash", "-c", "echo 40"})
	Cfg = s

	JSON := GetJSON()
	a.Contains(JSON, "processor_info")
	a.Contains(JSON, "temp")
	a.Contains(JSON, "usage")
	a.Contains(JSON, "load")

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(JSON), &d)
	a.Equal(err, nil)

}

func (a ApiAllSuite) TestCalculateTemp() {
	for _, test := range []struct {
		usage    float64
		expected float64
	}{
		{
			usage:    5,
			expected: 40.5,
		},
		{
			usage:    20,
			expected: 40.2,
		},
	} {
		a.Equal(test.expected, calculateTemp(40, test.usage))
	}
}

func (a ApiAllSuite) TestGetTempByUsage() {
	c := CPU{}
	for _, test := range []struct {
		percent  int
		expected float64
	}{
		{
			percent:  5,
			expected: 48.5,
		},
		{
			percent:  10,
			expected: 52.1,
		},
		{
			percent:  20,
			expected: 55.2,
		},
		{
			percent:  30,
			expected: 63.3,
		},
		{
			percent:  40,
			expected: 65.4,
		},
		{
			percent:  50,
			expected: 67.5,
		},
		{
			percent:  60,
			expected: 69.6,
		},
		{
			percent:  70,
			expected: 72.7,
		},
		{
			percent:  75,
			expected: 72.8,
		},
	} {
		c.ProcessorInfo.Usage.Percent = test.percent
		c.getTempByUsage()
		a.Equal(test.expected, c.ProcessorInfo.Temp.Actual)
	}
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
