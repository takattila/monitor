package network

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg           *settings.Settings
	Sleep         = 2 * time.Second
	L             logger.Logger
	netIOCounters = net.IOCounters
	netInterfaces = net.Interfaces
)

// Measurement holds a network counter sample and the time it was recorded.
type Measurement struct {
	BytesRecv  uint64
	BytesSent  uint64
	RecordedAt time.Time
}

// Measurements holds the latest counter sample of each interface.
var (
	Measurements = map[string]Measurement{}
	mu           sync.RWMutex
)

// StopStats stops the Stats background loop. Sending one value on it makes
// Stats return after the current iteration.
var StopStats = make(chan struct{})

// Stats keeps sampling the network counters in the background.
func Stats() {
	for {
		select {
		case <-StopStats:
			return
		default:
		}
		sample()
		time.Sleep(Sleep)
	}
}

// sample records the current network counters of each interface into
// Measurements, but only when the NetworkTraffic toggle is enabled.
func sample() {
	if Cfg.Data.GetBool("NetworkTraffic") {
		c, err := netIOCounters(true)
		L.Error(err)
		now := time.Now()

		mu.Lock()
		for _, n := range c {
			Measurements[n.Name] = Measurement{
				BytesRecv:  n.BytesRecv,
				BytesSent:  n.BytesSent,
				RecordedAt: now,
			}
		}
		mu.Unlock()
	}
}

// getMeasurement returns the recorded measurement of an interface, if any.
func getMeasurement(name string) (Measurement, bool) {
	mu.RLock()
	defer mu.RUnlock()
	start, ok := Measurements[name]
	return start, ok
}

// GetJSON returns with a JSON that holds information from network Traffic, from all interfaces.
func GetJSON() string {
	jsonArray := makeEmptyArray()
	if Cfg.Data.GetBool("NetworkTraffic") {
		c, err := netIOCounters(true)
		L.Error(err)
		for _, n := range c {
			start, ok := getMeasurement(n.Name)
			if ok && start.BytesRecv != 0 && start.BytesSent != 0 && !start.RecordedAt.IsZero() {
				in := computeRate(start.BytesRecv, n.BytesRecv, start.RecordedAt)
				out := computeRate(start.BytesSent, n.BytesSent, start.RecordedAt)
				jsonArray = append(jsonArray,
					`"`+fmt.Sprint(n.Name)+`": {
						"in": `+fmt.Sprint(in)+`,
						"out": `+fmt.Sprint(out)+`
					}
				`)
			}
		}
	}

	return `{ "network_info": {` + strings.Join(jsonArray, ",") + `}}`
}

// computeRate returns the traffic rate in KB/s between a counter sample and a
// newer counter value. It returns 0 when the elapsed time is not positive.
func computeRate(startByte, end uint64, recordedAt time.Time) float64 {
	elapsed := time.Since(recordedAt).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return round(float64(end-startByte) / 1024 / elapsed)
}

// round rounds a float to two decimal places.
func round(v float64) float64 {
	return math.Round(v*100) / 100
}

// makeEmptyArray creates an empty network array.
func makeEmptyArray() []string {
	jsonArray := make([]string, 0)
	c, err := netInterfaces()
	L.Error(err)
	for _, n := range c {
		jsonArray = append(jsonArray,
			`"`+fmt.Sprint(n.Name)+`": {
				"in": 0,
				"out": 0
			}
		`)
	}
	return jsonArray
}
