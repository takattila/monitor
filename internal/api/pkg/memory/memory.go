package memory

import (
	"encoding/json"

	"github.com/shirou/gopsutil/mem"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
)

type Mem struct {
	MemoryInfo struct {
		Total struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"total"`
		Used struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"used"`
		Free struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"free"`
		Cached struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"cached"`
		Available struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"available"`
		Swap struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"swap"`
		Video struct {
			Total      float64 `json:"total"`
			TotalUnit  string  `json:"total_unit"`
			Actual     float64 `json:"actual"`
			ActualUnit string  `json:"actual_unit"`
			Percent    float64 `json:"percent"`
		} `json:"video"`
	} `json:"memory_info"`
}

func GetJSON() string {
	m := Mem{}
	if Cfg.Data.GetBool("Memory") {
		vm, err := mem.VirtualMemory()
		common.Error(err)
		swp, err := mem.SwapMemory()
		common.Error(err)

		m.getTotal(vm)
		m.getUsed(vm)
		m.getFree(vm)
		m.getCached(vm)
		m.getAvailable(vm)
		m.getSwap(swp)
		m.getVideo(vm)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	common.Error(err)

	return string(b)
}

func getPhysicalMemory(vm *mem.VirtualMemoryStat) uint64 {
	ret, _ := Cfg.GetString("on_runtime.physical_memory")
	if ret == "" {
		return vm.Total
	}
	return common.TextToBytes(ret)
}

func (m *Mem) getTotal(vm *mem.VirtualMemoryStat) *Mem {
	PhysicalMemoryInBytes := getPhysicalMemory(vm)
	size_physical := common.DynamicSizeIECSize(PhysicalMemoryInBytes)
	unit_physical := common.DynamicSizeIECUnit(PhysicalMemoryInBytes)
	size_total := common.DynamicSizeIECSize(vm.Total)
	unit_total := common.DynamicSizeIECUnit(vm.Total)

	m.MemoryInfo.Total.Total = size_physical
	m.MemoryInfo.Total.TotalUnit = unit_physical
	m.MemoryInfo.Total.Actual = size_total
	m.MemoryInfo.Total.ActualUnit = unit_total
	m.MemoryInfo.Total.Percent = common.GetPercent(vm.Total, PhysicalMemoryInBytes)

	return m
}

func (m *Mem) getUsed(vm *mem.VirtualMemoryStat) *Mem {
	m.MemoryInfo.Used.Total = common.DynamicSizeIECSize(vm.Total)
	m.MemoryInfo.Used.TotalUnit = common.DynamicSizeIECUnit(vm.Total)
	m.MemoryInfo.Used.Actual = common.DynamicSizeIECSize(vm.Used)
	m.MemoryInfo.Used.ActualUnit = common.DynamicSizeIECUnit(vm.Used)
	m.MemoryInfo.Used.Percent = common.GetPercent(vm.Used, vm.Total)

	return m
}

func (m *Mem) getFree(vm *mem.VirtualMemoryStat) *Mem {
	m.MemoryInfo.Free.Total = common.DynamicSizeIECSize(vm.Total)
	m.MemoryInfo.Free.TotalUnit = common.DynamicSizeIECUnit(vm.Total)
	m.MemoryInfo.Free.Actual = common.DynamicSizeIECSize(vm.Free)
	m.MemoryInfo.Free.ActualUnit = common.DynamicSizeIECUnit(vm.Free)
	m.MemoryInfo.Free.Percent = common.GetPercent(vm.Free, vm.Total)

	return m
}

func (m *Mem) getCached(vm *mem.VirtualMemoryStat) *Mem {
	m.MemoryInfo.Cached.Total = common.DynamicSizeIECSize(vm.Total)
	m.MemoryInfo.Cached.TotalUnit = common.DynamicSizeIECUnit(vm.Total)
	m.MemoryInfo.Cached.Actual = common.DynamicSizeIECSize(vm.Cached)
	m.MemoryInfo.Cached.ActualUnit = common.DynamicSizeIECUnit(vm.Cached)
	m.MemoryInfo.Cached.Percent = common.GetPercent(vm.Cached, vm.Total)

	return m
}

func (m *Mem) getAvailable(vm *mem.VirtualMemoryStat) *Mem {
	m.MemoryInfo.Available.Total = common.DynamicSizeIECSize(vm.Total)
	m.MemoryInfo.Available.TotalUnit = common.DynamicSizeIECUnit(vm.Total)
	m.MemoryInfo.Available.Actual = common.DynamicSizeIECSize(vm.Available)
	m.MemoryInfo.Available.ActualUnit = common.DynamicSizeIECUnit(vm.Available)
	m.MemoryInfo.Available.Percent = common.GetPercent(vm.Available, vm.Total)

	return m
}

func (m *Mem) getSwap(swp *mem.SwapMemoryStat) *Mem {
	m.MemoryInfo.Swap.Total = common.DynamicSizeIECSize(swp.Total)
	m.MemoryInfo.Swap.TotalUnit = common.DynamicSizeIECUnit(swp.Total)
	m.MemoryInfo.Swap.Actual = common.DynamicSizeIECSize(swp.Used)
	m.MemoryInfo.Swap.ActualUnit = common.DynamicSizeIECUnit(swp.Used)
	m.MemoryInfo.Swap.Percent = common.GetPercent(swp.Used, swp.Total)

	return m
}

func (m *Mem) getVideo(vm *mem.VirtualMemoryStat) *Mem {
	PhysicalMemoryInBytes := getPhysicalMemory(vm)
	videoMemory := PhysicalMemoryInBytes - vm.Total
	m.MemoryInfo.Video.Total = common.DynamicSizeIECSize(PhysicalMemoryInBytes)
	m.MemoryInfo.Video.TotalUnit = common.DynamicSizeIECUnit(PhysicalMemoryInBytes)
	m.MemoryInfo.Video.Actual = common.DynamicSizeIECSize(videoMemory)
	m.MemoryInfo.Video.ActualUnit = common.DynamicSizeIECUnit(videoMemory)
	m.MemoryInfo.Video.Percent = common.GetPercent(videoMemory, PhysicalMemoryInBytes)

	return m
}
