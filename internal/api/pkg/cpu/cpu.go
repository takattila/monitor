package cpu

import (
	"encoding/json"
	"math"
	"regexp"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
)

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

func GetJSON() string {
	c := CPU{}
	c.getUsage()
	c.getTemp()
	c.getLoad()

	b, err := json.Marshal(c)
	common.Error(err)

	return string(b)
}

func (c *CPU) getUsage() *CPU {
	percentage, err := cpu.Percent(0, true)
	common.Error(err)

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
		common.Error(err)

		c.ProcessorInfo.Temp.Total = 100
		c.ProcessorInfo.Temp.TotalUnit = "°C"
		c.ProcessorInfo.Temp.Actual = temp
		c.ProcessorInfo.Temp.ActualUnit = "°C"
		c.ProcessorInfo.Temp.Percent = temp
	}
	return c
}

func (c *CPU) getLoad() *CPU {
	load, err := load.Avg()
	common.Error(err)

	c.ProcessorInfo.Load.Min01 = load.Load1
	c.ProcessorInfo.Load.Min05 = load.Load5
	c.ProcessorInfo.Load.Min15 = load.Load15

	return c
}
