package memory

import (
	"encoding/json"
	"errors"
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

func (a ApiMemorySuite) TestGetMemoryFromConfigSuccess() {
	s := getConfig("api", "linux")

	memStruct := Mem{}
	memStruct.MemoryInfo.Total.Total = 1
	memStruct.MemoryInfo.Total.TotalUnit = "GB"
	memStruct.MemoryInfo.Total.Actual = 0.9
	memStruct.MemoryInfo.Total.ActualUnit = "GB"
	memStruct.MemoryInfo.Total.Percent = 90

	jsonBytes, err := json.Marshal(memStruct)
	a.Nil(err)

	s.Data.Set("on_runtime.memory", []string{"echo", string(jsonBytes)})
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	result, err := getMemoryFromConfig()
	a.Nil(err)
	a.NotNil(result)
	a.Equal(1.0, result.MemoryInfo.Total.Total)
	a.Equal("GB", result.MemoryInfo.Total.TotalUnit)
}

func (a ApiMemorySuite) TestGetMemoryFromConfigJsonUnmarshalError() {
	s := getConfig("api", "linux")

	memStruct := Mem{}
	memStruct.MemoryInfo.Total.Total = 1
	memStruct.MemoryInfo.Total.TotalUnit = "GB"
	memStruct.MemoryInfo.Total.Actual = 0.9
	memStruct.MemoryInfo.Total.ActualUnit = "GB"
	memStruct.MemoryInfo.Total.Percent = 90

	jsonBytes, err := json.Marshal(memStruct)
	a.Nil(err)

	s.Data.Set("on_runtime.memory", []string{"echo", string(jsonBytes)})
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	// Set up the mock function *before* calling the function under test
	oldJsonUnmarshal := jsonUnmarshal
	jsonUnmarshal = func(data []byte, v any) error {
		return errors.New("invalid character 'e' looking for beginning of value")
	}
	// Defer the restoration of the original function
	defer func() { jsonUnmarshal = oldJsonUnmarshal }()

	// Now, call the function under test
	result, err := getMemoryFromConfig()

	// Assert that an error was returned and the result is nil
	a.Assertions.Nil(result)
	a.Assertions.Error(err)
	a.Contains(err.Error(), "invalid character")
}

func (a ApiMemorySuite) TestGetMemoryFromConfigInvalidJSON() {
	s := getConfig("api", "linux")

	s.Data.Set("on_runtime.memory", []string{"echo", "not-a-json"})
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	result, err := getMemoryFromConfig()
	a.Nil(result)
	a.Error(err)
	a.Contains(err.Error(), "invalid character")
}

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

func (a ApiMemorySuite) TestGetTotal() {
	s := getConfig("api", "linux")
	s.Data.Set("on_runtime.physical_memory", "16GB")
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getTotal(vm)

	a.Greater(m.MemoryInfo.Total.Total, float64(0))
	a.NotEmpty(m.MemoryInfo.Total.TotalUnit)
	a.Greater(m.MemoryInfo.Total.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Total.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Total.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetUsed() {
	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getUsed(vm)

	a.Greater(m.MemoryInfo.Used.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Used.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Used.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetFree() {
	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getFree(vm)

	a.GreaterOrEqual(m.MemoryInfo.Free.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Free.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Free.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetCached() {
	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getCached(vm)

	a.GreaterOrEqual(m.MemoryInfo.Cached.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Cached.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Cached.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetAvailable() {
	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getAvailable(vm)

	a.GreaterOrEqual(m.MemoryInfo.Available.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Available.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Available.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetSwap() {
	swp, err := mem.SwapMemory()
	a.Nil(err)

	m := Mem{}
	m.getSwap(swp)

	a.GreaterOrEqual(m.MemoryInfo.Swap.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Swap.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Swap.Percent, float64(0))
}

func (a ApiMemorySuite) TestGetVideo() {
	s := getConfig("api", "linux")
	s.Data.Set("on_runtime.physical_memory", "16GB")
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	vm, err := mem.VirtualMemory()
	a.Nil(err)

	m := Mem{}
	m.getVideo(vm)

	a.GreaterOrEqual(m.MemoryInfo.Video.Actual, float64(0))
	a.NotEmpty(m.MemoryInfo.Video.ActualUnit)
	a.GreaterOrEqual(m.MemoryInfo.Video.Percent, float64(0))
}

func (a *ApiMemorySuite) TestGetJSONFromConfig() {
	s := getConfig("api", "linux")

	memStruct := Mem{}
	memStruct.MemoryInfo.Total.Total = 1.0
	memStruct.MemoryInfo.Total.TotalUnit = "GB"
	memStruct.MemoryInfo.Total.Actual = 0.9
	memStruct.MemoryInfo.Total.ActualUnit = "GB"
	memStruct.MemoryInfo.Total.Percent = 90.0

	jsonBytes, err := json.Marshal(memStruct)
	a.Nil(err)

	s.Data.Set("on_runtime.memory", []string{"echo", string(jsonBytes)})
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	oldJsonUnmarshal := jsonUnmarshal
	jsonUnmarshal = func(data []byte, v any) error {
		return oldJsonUnmarshal(data, v)
	}
	defer func() { jsonUnmarshal = oldJsonUnmarshal }()

	resultJSON := GetJSON()

	a.NotEqual("{}", resultJSON)

	expectedJSON, err := json.MarshalIndent(memStruct, "", "  ")
	a.Nil(err)

	a.Equal(string(expectedJSON), resultJSON)
}

func (a *ApiMemorySuite) TestGetJSONVirtualMemoryError() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", true)
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	oldVirtualMemory := memVirtualMemory
	memVirtualMemory = func() (*mem.VirtualMemoryStat, error) {
		return nil, errors.New("mocked VirtualMemory error")
	}
	defer func() { memVirtualMemory = oldVirtualMemory }()

	oldSwapMemory := memSwapMemory
	memSwapMemory = func() (*mem.SwapMemoryStat, error) {
		return &mem.SwapMemoryStat{}, nil
	}
	defer func() { memSwapMemory = oldSwapMemory }()

	result := GetJSON()

	a.Equal("{}", result)
}

func (a *ApiMemorySuite) TestGetJSONSwapMemoryError() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", true)
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	oldMemVirtualMemory := memVirtualMemory
	memVirtualMemory = func() (*mem.VirtualMemoryStat, error) {
		return &mem.VirtualMemoryStat{}, nil
	}
	defer func() { memVirtualMemory = oldMemVirtualMemory }()

	oldMemSwapMemory := memSwapMemory
	memSwapMemory = func() (*mem.SwapMemoryStat, error) {
		return nil, errors.New("mocked SwapMemory error")
	}
	defer func() { memSwapMemory = oldMemSwapMemory }()

	result := GetJSON()

	a.Equal("{}", result)
}

func (a *ApiMemorySuite) TestGetJSONMarshalError() {
	s := getConfig("api", "linux")
	memStruct := Mem{}
	jsonBytes, err := json.Marshal(memStruct)
	a.Nil(err)
	s.Data.Set("on_runtime.memory", []string{"echo", string(jsonBytes)})
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)

	oldJsonMarshalIndent := jsonMarshalIndent
	jsonMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("mocked json.MarshalIndent error")
	}
	defer func() { jsonMarshalIndent = oldJsonMarshalIndent }()

	result := GetJSON()

	a.Equal("{}", result)
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
