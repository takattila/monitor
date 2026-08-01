package network

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	ApiNetworkSuite struct {
		suite.Suite
	}
)

func (a ApiNetworkSuite) setup(toggle bool) *settings.Settings {
	s := getConfig("api", "linux")
	s.Data.Set("NetworkTraffic", toggle)
	Cfg = s
	L = logger.New(logger.NoneLevel, logger.ColorOff)
	mu.Lock()
	Measurements = map[string]Measurement{}
	mu.Unlock()
	return s
}

func (a ApiNetworkSuite) mockIOCounters(stats []net.IOCountersStat, err error) {
	netIOCounters = func(bool) ([]net.IOCountersStat, error) {
		return stats, err
	}
}

func (a ApiNetworkSuite) mockInterfaces(ifaces []net.InterfaceStat, err error) {
	netInterfaces = func() ([]net.InterfaceStat, error) {
		return ifaces, err
	}
}

func (a ApiNetworkSuite) TestGetJSONToggleOff() {
	s := a.setup(false)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{{Name: "eth0"}, {Name: "wlan0"}}, nil)
	a.mockIOCounters(nil, nil)

	JSON := GetJSON()
	a.Contains(JSON, "network_info")
	a.Contains(JSON, "eth0")
	a.Contains(JSON, "wlan0")

	d := make(map[string]interface{})
	a.Nil(json.Unmarshal([]byte(JSON), &d))
}

func (a ApiNetworkSuite) TestGetJSONToggleOffInterfacesError() {
	s := a.setup(false)
	_ = s

	a.mockInterfaces(nil, fmt.Errorf("interfaces error"))
	a.mockIOCounters(nil, nil)

	JSON := GetJSON()
	a.Contains(JSON, "network_info")

	d := make(map[string]interface{})
	a.Nil(json.Unmarshal([]byte(JSON), &d))
}

func (a ApiNetworkSuite) TestGetJSONToggleOffNoInterfaces() {
	s := a.setup(false)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{}, nil)
	a.mockIOCounters(nil, nil)

	a.Equal(`{ "network_info": {}}`, GetJSON())
}

func (a ApiNetworkSuite) TestGetJSONToggleOnIOCountersError() {
	s := a.setup(true)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{{Name: "eth0"}}, nil)
	a.mockIOCounters(nil, fmt.Errorf("io counters error"))

	JSON := GetJSON()
	a.Contains(JSON, "network_info")
	a.Contains(JSON, "eth0")
	a.Contains(JSON, `"in": 0`)

	d := make(map[string]interface{})
	a.Nil(json.Unmarshal([]byte(JSON), &d))
}

func (a ApiNetworkSuite) TestGetJSONToggleOnWithoutCounters() {
	s := a.setup(true)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{{Name: "eth0"}}, nil)
	a.mockIOCounters([]net.IOCountersStat{}, nil)

	JSON := GetJSON()
	a.Contains(JSON, "eth0")
	a.Contains(JSON, `"in": 0`)
	a.Contains(JSON, `"out": 0`)
}

func (a ApiNetworkSuite) TestGetJSONToggleOnWithoutMeasurement() {
	s := a.setup(true)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{}, nil)
	a.mockIOCounters([]net.IOCountersStat{{Name: "eth0", BytesRecv: 100, BytesSent: 200}}, nil)

	a.Equal(`{ "network_info": {}}`, GetJSON())
}

func (a ApiNetworkSuite) TestGetJSONToggleOnInvalidMeasurements() {
	s := a.setup(true)
	_ = s
	a.mockInterfaces([]net.InterfaceStat{}, nil)

	for _, test := range []struct {
		name string
		meas Measurement
	}{
		{
			name: "BytesRecv is zero",
			meas: Measurement{BytesRecv: 0, BytesSent: 200, RecordedAt: time.Now()},
		},
		{
			name: "BytesSent is zero",
			meas: Measurement{BytesRecv: 100, BytesSent: 0, RecordedAt: time.Now()},
		},
		{
			name: "RecordedAt is zero",
			meas: Measurement{BytesRecv: 100, BytesSent: 200, RecordedAt: time.Time{}},
		},
	} {
		Measurements["eth0"] = test.meas
		a.mockIOCounters([]net.IOCountersStat{{Name: "eth0", BytesRecv: 100, BytesSent: 200}}, nil)

		a.Equal(`{ "network_info": {}}`, GetJSON(), test.name)
	}
}

func (a ApiNetworkSuite) TestGetJSONToggleOnValidMeasurement() {
	s := a.setup(true)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{}, nil)
	a.mockIOCounters([]net.IOCountersStat{{Name: "eth0", BytesRecv: 3048, BytesSent: 4069}}, nil)

	Measurements["eth0"] = Measurement{
		BytesRecv:  1000,
		BytesSent:  2000,
		RecordedAt: time.Now().Add(-2 * time.Second),
	}

	JSON := GetJSON()
	a.Contains(JSON, "eth0")
	a.Contains(JSON, "in")
	a.Contains(JSON, "out")

	d := map[string]map[string]map[string]float64{}
	a.Nil(json.Unmarshal([]byte(JSON), &d))

	a.InDelta(1.0, d["network_info"]["eth0"]["in"], 0.1)
	a.InDelta(1.01, d["network_info"]["eth0"]["out"], 0.1)
}

func (a ApiNetworkSuite) TestGetJSONToggleOnFutureMeasurement() {
	s := a.setup(true)
	_ = s

	a.mockInterfaces([]net.InterfaceStat{}, nil)
	a.mockIOCounters([]net.IOCountersStat{{Name: "eth0", BytesRecv: 100, BytesSent: 200}}, nil)

	Measurements["eth0"] = Measurement{
		BytesRecv:  1000,
		BytesSent:  2000,
		RecordedAt: time.Now().Add(time.Hour),
	}

	JSON := GetJSON()
	a.Contains(JSON, "eth0")
	a.Contains(JSON, `"in": 0`)
	a.Contains(JSON, `"out": 0`)
}

func (a ApiNetworkSuite) TestSampleToggleOn() {
	s := a.setup(true)
	_ = s

	a.mockIOCounters([]net.IOCountersStat{
		{Name: "eth0", BytesRecv: 1000, BytesSent: 2000},
		{Name: "wlan0", BytesRecv: 3000, BytesSent: 4000},
	}, nil)

	sample()

	a.Len(Measurements, 2)
	a.Contains(Measurements, "eth0")
	a.Contains(Measurements, "wlan0")
	a.Equal(uint64(1000), Measurements["eth0"].BytesRecv)
	a.Equal(uint64(2000), Measurements["eth0"].BytesSent)
	a.False(Measurements["eth0"].RecordedAt.IsZero())
}

func (a ApiNetworkSuite) TestSampleToggleOnEmptyCounters() {
	s := a.setup(true)
	_ = s

	a.mockIOCounters([]net.IOCountersStat{}, nil)

	sample()

	a.Len(Measurements, 0)
}

func (a ApiNetworkSuite) TestSampleToggleOnError() {
	s := a.setup(true)
	_ = s

	a.mockIOCounters(nil, fmt.Errorf("io counters error"))

	sample()

	a.Len(Measurements, 0)
}

func (a ApiNetworkSuite) TestSampleToggleOff() {
	s := a.setup(false)
	_ = s

	a.mockIOCounters([]net.IOCountersStat{
		{Name: "eth0", BytesRecv: 1000, BytesSent: 2000},
	}, nil)

	sample()

	a.Len(Measurements, 0)
}

func (a ApiNetworkSuite) TestStats() {
	s := a.setup(true)
	_ = s

	a.mockIOCounters([]net.IOCountersStat{
		{Name: "eth0", BytesRecv: 1000, BytesSent: 2000},
	}, nil)

	Sleep = 5 * time.Millisecond
	go Stats()

	time.Sleep(30 * time.Millisecond)

	StopStats <- struct{}{}
	time.Sleep(10 * time.Millisecond)

	mu.RLock()
	a.NotEmpty(Measurements)
	a.Contains(Measurements, "eth0")
	mu.RUnlock()
}

func (a ApiNetworkSuite) TestComputeRate() {
	recordedAt := time.Now().Add(-2 * time.Second)

	a.InDelta(1.0, computeRate(1000, 3048, recordedAt), 0.1)
	a.InDelta(1.01, computeRate(2000, 4069, recordedAt), 0.1)

	a.Equal(0.0, computeRate(1000, 3048, time.Now().Add(time.Hour)))
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
