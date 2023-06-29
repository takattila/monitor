package processes

import (
	"fmt"
	"math"
	"strings"

	"github.com/bradfitz/slice"
	"github.com/shirou/gopsutil/process"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
	L   logger.Logger
)

type (
	Process struct {
		Pid           int32
		Name          string
		User          string
		MemoryPercent float64
		CPUPercent    float64
		Cmdline       string
	}

	Processes []Process
)

// GetJSON returns with a JSON that holds information from processes that stored in the configuration.
func GetJSON() string {
	var out string
	jsonArray := make([]string, 0)
	ret, _ := Cfg.GetStringSlice("on_runtime.commands.processes")

	if Cfg.Data.GetBool("TopProcesses") {
		if len(ret) == 0 {
			p := Processes{}
			return p.getAllProcesses().orderByCPUPercent().getFirst10Process().getJSON()
		} else {
			out = common.Cli(ret)
		}
	}

	processLines := strings.Split(out, "\n")

	if len(processLines) > 1 {
		if Cfg.Data.GetBool("TopProcesses") {
			i := 0
			for _, line := range processLines {
				line = strings.Join(strings.Fields(line), " ")
				line = strings.ReplaceAll(line, `"`, `'`)
				line = strings.ReplaceAll(line, `\`, `\\`)
				if line != "" {
					i++
					jsonArray = append(jsonArray, `"`+fmt.Sprint(i)+`": {
							"pid": "`+getPid(line)+`",
							"user": "`+getUser(line)+`",
							"mem": "`+getMem(line)+`",
							"cpu": "`+getCpu(line)+`",
							"cmd": "`+getCmd(line)+`"
						}
					`)
				}
			}
		}
	} else {
		jsonArray = append(jsonArray, `"1": {
					"pid": "1",
					"user": "root",
					"mem": "0.0%",
					"cpu": "0.0%",
					"cmd": "/sbin/init"
				}
			`)
	}

	return `{ "process_info": {` + strings.Join(jsonArray, ",") + `}}`
}

// getPid fetches PID from string.
func getPid(s string) string {
	res := strings.Split(s, " ")
	if (s == "") || (s == " ") || (len(res) == 1 && s == "") {
		return "0"
	}
	return res[0]
}

// getUser fetches user from string.
func getUser(s string) string {
	res := strings.Split(s, " ")
	if len(res) < 2 {
		return "unknown"
	}
	return res[1]
}

// getMem fetches mem from string.
func getMem(s string) string {
	res := strings.Split(s, " ")
	if len(res) < 3 {
		return "0.0%"
	}
	return res[2]
}

// getCpu fetches CPU from string.
func getCpu(s string) string {
	res := strings.Split(s, " ")
	if len(res) < 4 {
		return "0.0%"
	}
	return res[3]
}

// getCmd fetches command from string.
func getCmd(s string) string {
	array := strings.Split(s, " ")
	res := "unknown"
	if len(array) > 4 {
		res = strings.Join(append(array[4:]), " ")
	}
	return res
}

// escapeStr escapes special characters that breaks JSON creation.
func escapeStr(str string) string {
	str = strings.ReplaceAll(str, `"`, "'")
	str = strings.ReplaceAll(str, `\`, "/")
	return str
}

// getAllProcesses makes a Processes structure.
func (p *Processes) getAllProcesses() *Processes {
	processes, err := process.Processes()
	L.Error(err)

	var filtered Processes

	for _, p := range processes {
		decimalPlaces := 1
		ratio := math.Pow(10, float64(decimalPlaces))

		pId := p.Pid
		pName, err := p.Name()
		L.Error(err)
		pName = escapeStr(pName)

		pUser, err := p.Username()
		L.Error(err)
		pUser = escapeStr(pUser)

		m, err := p.MemoryPercent()
		L.Error(err)
		pMem := math.Round(float64(m)*ratio) / ratio

		pCpu, err := p.CPUPercent()
		L.Error(err)
		pCpu = math.Round(pCpu*ratio) / ratio

		pCmdline, err := p.Cmdline()
		L.Error(err)
		pCmdline = escapeStr(pCmdline)

		if pName != "" && !strings.Contains(pName, "cmd/api") {
			filtered = append(filtered, Process{
				Pid:           pId,
				Name:          pName,
				User:          pUser,
				MemoryPercent: pMem,
				CPUPercent:    pCpu,
				Cmdline:       pCmdline,
			})
		}
	}

	return &filtered
}

// orderByCPUPercent orders the Processes struct by CPU percentage.
func (p *Processes) orderByCPUPercent() *Processes {
	processes := *p

	slice.Sort(processes[:], func(i, j int) bool {
		return processes[i].CPUPercent > processes[j].CPUPercent
	})

	return &processes
}

// getFirst10Process returns with the first top 10 processes information.
func (p *Processes) getFirst10Process() *Processes {
	processes := *p
	var filtered Processes

	i := 0
	for _, proc := range processes {
		i++
		if i <= 10 {
			filtered = append(filtered, Process{
				Pid:           proc.Pid,
				User:          proc.User,
				MemoryPercent: proc.MemoryPercent,
				CPUPercent:    proc.CPUPercent,
				Cmdline:       proc.Cmdline,
			})
		}
	}
	return &filtered
}

// getJSON makes a JSON object from the Processes struct.
func (p *Processes) getJSON() string {
	var lines []string

	i := 0
	for _, proc := range *p {
		i++
		if i <= 10 {
			lines = append(lines, `"`+fmt.Sprint(i)+`": {
					"pid": "`+fmt.Sprint(proc.Pid)+`",
					"user": "`+proc.User+`",
					"mem": "`+fmt.Sprint(proc.MemoryPercent)+`",
					"cpu": "`+fmt.Sprint(proc.CPUPercent)+`",
					"cmd": "`+proc.Cmdline+`"
				}
			`)
		}
	}
	return `{ "process_info": {` + strings.Join(lines, ",") + `}}`
}
