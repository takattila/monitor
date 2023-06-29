package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg            *settings.Settings
	StartMesureIn  = map[string]uint64{}
	StartMesureOut = map[string]uint64{}
	Sleep          = 2 * time.Second
	L              logger.Logger
)

// Stats collects network data into StartMesureIn and StartMesureOut variables.
// It should be run in the background by starting with: 'go Stats()'.
func Stats() {
	for {
		if Cfg.Data.GetBool("NetworkTraffic") {
			c, err := net.IOCounters(true)
			L.Error(err)
			for _, n := range c {
				StartMesureIn[n.Name] = n.BytesRecv
				StartMesureOut[n.Name] = n.BytesSent
			}
		}
		time.Sleep(Sleep)
	}
}

// GetJSON returns with a JSON that holds information from network Traffic, from all interfaces.
func GetJSON() string {
	jsonArray := makeEmptyArray()
	if Cfg.Data.GetBool("NetworkTraffic") {
		c, err := net.IOCounters(true)
		L.Error(err)
		for _, n := range c {
			if StartMesureIn[n.Name] != 0 && StartMesureOut[n.Name] != 0 {
				endMesureIn := n.BytesRecv
				endMesureOut := n.BytesSent

				in := (endMesureIn - StartMesureIn[n.Name]) / 1024
				out := (endMesureOut - StartMesureOut[n.Name]) / 1024

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

// makeEmptyArray creates an empty network array.
func makeEmptyArray() []string {
	jsonArray := make([]string, 0)
	c, err := net.Interfaces()
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
