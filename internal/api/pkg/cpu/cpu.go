package cpu

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
	L   logger.Logger
)

// The CPU structure contains the necessary data about the CPU.
type CPU struct {
	ProcessorInfo struct {
		Temp struct {
			Total      int     `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"temp"`
		Usage struct {
			Total      int    `json:"total"`
			TotalUnit  string `json:"total_unit"`
			Actual     int    `json:"actual"`
			ActualUnit string `json:"actual_unit"`
			Percent    int    `json:"percent"`
		} `json:"usage"`
		Load struct {
			Min01 float64 `json:"min_01"`
			Min05 float64 `json:"min_05"`
			Min15 float64 `json:"min_15"`
		} `json:"load"`
	} `json:"processor_info"`
}

// GetJSON returns with a JSON that holds all necessary CPU information.
func GetJSON() string {
	c := CPU{}
	c.getUsage()
	c.getTemp()
	c.getLoad()

	b, err := json.Marshal(c)
	L.Error(err)

	return string(b)
}

// getUsage populates the usage data.
func (c *CPU) getUsage() *CPU {
	percentage, err := cpu.Percent(0, true)
	L.Error(err)

	var percents float64
	var numberOfCores int

	for _, cpupercent := range percentage {
		percents += cpupercent
		numberOfCores++
	}

	percentAll := int(percents / float64(numberOfCores))

	c.ProcessorInfo.Usage.Total = 100
	c.ProcessorInfo.Usage.TotalUnit = "%"
	c.ProcessorInfo.Usage.Actual = percentAll
	c.ProcessorInfo.Usage.ActualUnit = "%"
	c.ProcessorInfo.Usage.Percent = percentAll

	return c
}

// calculateTemptJSON populates the temperature data.
func calculateTemp(temp, usage float64) float64 {
	decimalPlaces := 1
	divider := float64(0)

	if usage < 10 {
		divider = 10
	} else {
		divider = 100
	}

	result := temp + (usage / divider)
	ratio := math.Pow(10, float64(decimalPlaces))
	return math.Round(result*ratio) / ratio
}

// getTempByUsage calculates the approximate temperature from the CPU usage.
func (c *CPU) getTempByUsage() *CPU {
	temp := 0.1
	usage := float64(c.ProcessorInfo.Usage.Percent)

	if usage < 10 {
		temp = calculateTemp(48, usage)
	}
	if usage >= 10 && usage <= 20 {
		temp = calculateTemp(52, usage)
	}
	if usage >= 20 && usage <= 30 {
		temp = calculateTemp(55, usage)
	}
	if usage >= 30 && usage <= 40 {
		temp = calculateTemp(63, usage)
	}
	if usage >= 40 && usage <= 50 {
		temp = calculateTemp(65, usage)
	}
	if usage >= 50 && usage <= 60 {
		temp = calculateTemp(67, usage)
	}
	if usage >= 60 && usage <= 70 {
		temp = calculateTemp(69, usage)
	}
	if usage >= 70 {
		temp = calculateTemp(72, usage)
	}

	c.ProcessorInfo.Temp.Total = 100
	c.ProcessorInfo.Temp.TotalUnit = "°C"
	c.ProcessorInfo.Temp.Actual = temp
	c.ProcessorInfo.Temp.ActualUnit = "°C"
	c.ProcessorInfo.Temp.Percent = temp

	return c
}

// getTemp fetches the CPU temperature by running a command.
func (c *CPU) getTemp() *CPU {
	ret, _ := Cfg.GetStringSlice("on_runtime.commands.cpu_temp")
	if len(ret) == 0 {
		return c.getTempByUsage()
	}
	res := common.Cli(ret)
	if res != "" {
		re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
		floatStr := re.FindAllString(res, -1)[0]
		temp, err := strconv.ParseFloat(floatStr, 64)
		L.Error(err)

		c.ProcessorInfo.Temp.Total = 100
		c.ProcessorInfo.Temp.TotalUnit = "°C"
		c.ProcessorInfo.Temp.Actual = temp
		c.ProcessorInfo.Temp.ActualUnit = "°C"
		c.ProcessorInfo.Temp.Percent = temp
	}
	return c
}

// getLoad populates the CPU loads: 1, 5 or 15.
func (c *CPU) getLoad() *CPU {
	load, err := load.Avg()
	L.Error(err)

	c.ProcessorInfo.Load.Min01 = load.Load1
	c.ProcessorInfo.Load.Min05 = load.Load5
	c.ProcessorInfo.Load.Min15 = load.Load15

	return c
}
