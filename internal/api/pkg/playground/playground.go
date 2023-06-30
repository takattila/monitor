package playground

import (
	"fmt"

	"github.com/shirou/gopsutil/host"
	"github.com/takattila/monitor/internal/api/pkg/uptime"
	"github.com/takattila/monitor/pkg/logger"
)

var (
	L logger.Logger
)

// Playground for testing stuff.
func Playground() string {
	style := `style="background-color:#0d1117;color:#ffffff;margin:20px"`
	html := `<html><head><title>Go Playground</title></head><body ` + style + `><h1>Go</h1>`

	html = html + uptime.GetJSON() + "<br>"

	p := GetPlatform()
	L.Error(p.Error)
	L.Info(fmt.Sprintf("platform: %s, family: %s, version: %s", p.Name, p.Family, p.Version))
	html = html + fmt.Sprintf("platform: %s, family: %s, version: %s", p.Name, p.Family, p.Version) + "<br>"

	t, err := GetTemp()
	L.Error(err)
	L.Info(fmt.Sprintf("temps: %s", t))
	html = html + fmt.Sprintf("temps: %s", t) + "<br>"

	html = html + "</body></html>"
	return html
}

type Platform struct {
	Name    string
	Family  string
	Version string
	Error   error
}

func GetPlatform() (p Platform) {
	platform, family, version, err := host.PlatformInformation()
	p.Error = err

	p.Name = platform
	p.Family = family
	p.Version = version

	return p
}

func GetTemp() ([]host.TemperatureStat, error) {
	return host.SensorsTemperatures()
}
