package network

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	"github.com/takattila/settings-manager"
)

var (
	Cfg            *settings.Settings
	StartMesureIn  = map[string]uint64{}
	StartMesureOut = map[string]uint64{}
)

func Stats() {
	for {
		if Cfg.Data.Get("NetworkTraffic") == "true" {
			c, _ := net.IOCounters(true)
			for _, n := range c {
				StartMesureIn[n.Name] = n.BytesRecv
				StartMesureOut[n.Name] = n.BytesSent
			}
		}
		time.Sleep(2 * time.Second)
	}
}

func GetJSON() string {
	jsonArray := makeEmptyArray()
	if Cfg.Data.GetBool("NetworkTraffic") {
		c, _ := net.IOCounters(true)
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

func makeEmptyArray() []string {
	jsonArray := make([]string, 0)
	c, _ := net.Interfaces()
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
