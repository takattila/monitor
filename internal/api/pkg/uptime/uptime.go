package uptime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/host"
	"github.com/takattila/monitor/pkg/logger"
)

var (
	L logger.Logger
)

// The Uptime structure holds the uptime information of a JSON object.
type Uptime struct {
	Info string `json:"uptime_info"`
}

// GetJSON provides an uptime JSON.
func GetJSON() string {
	info, err := GetUptime().String()
	L.Error(err)

	up := Uptime{}
	up.Info = info

	b, err := json.Marshal(up)
	L.Error(err)

	return string(b)
}

// Up holds the uptime information.
type Up struct {
	Years   uint64
	Months  uint64
	Weeks   uint64
	Days    uint64
	Hours   uint64
	Minutes uint64
	Seconds uint64
	Error   error
}

// GetUptime returns the uptime information.
func GetUptime() (up Up) {
	uptime, err := host.Uptime()

	years := uptime / 60 / 60 / 24 / 7 / 30 / 12
	seconds := uptime % (60 * 60 * 24 * 7 * 30 * 12)
	months := seconds / 60 / 60 / 24 / 7 / 30
	seconds = uptime % (60 * 60 * 24 * 7 * 30)
	weeks := seconds / 60 / 60 / 24 / 7
	seconds = uptime % (60 * 60 * 24 * 7)
	days := seconds / 60 / 60 / 24
	seconds = uptime % (60 * 60 * 24)
	hours := seconds / 60 / 60
	seconds = uptime % (60 * 60)
	minutes := seconds / 60
	seconds = uptime % 60

	return Up{
		Years:   years,
		Months:  months,
		Weeks:   weeks,
		Days:    days,
		Hours:   hours,
		Minutes: minutes,
		Seconds: seconds,
		Error:   err,
	}
}

// String returns the uptime information coonverted to a string.
func (up Up) String() (string, error) {
	if up.Error != nil {
		return "", up.Error
	}

	var years, months, weeks, days, hours, minutes, seconds string

	if up.Years > 0 {
		years = plural("year", up.Years)
	}
	if up.Months > 0 {
		months = plural("month", up.Months)
	}
	if up.Weeks > 0 {
		weeks = plural("week", up.Weeks)
	}
	if up.Days > 0 {
		days = plural("day", up.Days)
	}
	if up.Hours > 0 {
		hours = plural("hour", up.Hours)
	}
	if up.Minutes > 0 {
		minutes = plural("minute", up.Minutes)
	}
	if up.Seconds > 0 {
		seconds = plural("second", up.Seconds)
	}

	var uptimeSlice []string
	for _, elem := range []string{years, months, weeks, days, hours, minutes, seconds} {
		if elem != "" {
			uptimeSlice = append(uptimeSlice, elem)
		}
	}

	return strings.Join(uptimeSlice, ", "), nil
}

// plural prints uptime units in plural if the unit is greater than 1.
func plural(unit string, num uint64) string {
	if num > 1 {
		return fmt.Sprintf("%d %ss", num, unit)
	}
	return fmt.Sprintf("%d %s", num, unit)
}
