package playground

import (
	"fmt"

	"github.com/shirou/gopsutil/host"
	"github.com/takattila/monitor/pkg/logger"
)

// Playground for testing stuff.
func Playground() string {
	l := logger.New(logger.DebugLevel, logger.ColorOn)
	html := "<html><head><title>Go Playground</title></head><body><h1>Go</h1>"

	up := GetUptime()
	l.Error(up.Error)
	l.Info(fmt.Sprintf("%d days, %d hours, %d minutes", up.Days, up.Hours, up.Minutes))
	html = html + fmt.Sprintf("%d days, %d hours, %d minutes", up.Days, up.Hours, up.Minutes) + "<br>"

	p := GetPlatform()
	l.Error(p.Error)
	l.Info(fmt.Sprintf("platform: %s, family: %s, version: %s", p.Name, p.Family, p.Version))
	html = html + fmt.Sprintf("platform: %s, family: %s, version: %s", p.Name, p.Family, p.Version) + "<br>"

	t, err := GetTemp()
	l.Error(err)
	l.Info(fmt.Sprintf("temps: %s", t))
	html = html + fmt.Sprintf("temps: %s", t) + "<br>"

	html = html + "</body></html>"
	return html
}

type UpTime struct {
	Days    uint64
	Hours   uint64
	Minutes uint64
	Error   error
}

func GetUptime() (up UpTime) {
	uptime, err := host.Uptime()
	up.Error = err

	days := uptime / (60 * 60 * 24)
	hours := (uptime - (days * 60 * 60 * 24)) / (60 * 60)
	minutes := ((uptime - (days * 60 * 60 * 24)) - (hours * 60 * 60)) / 60

	up.Days = days
	up.Hours = hours
	up.Minutes = minutes

	return up
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
