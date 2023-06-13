package services

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var (
	Cfg    *settings.Settings
	output = ""
)

func Watcher() {
	ServicesList := config.GetStringSlice(Cfg, "on_runtime.services_list")
	output = getProcessesStatus(ServicesList)
	for {
		if Cfg.Data.GetBool("Services") {
			ServicesList := config.GetStringSlice(Cfg, "on_runtime.services_list")
			output = getProcessesStatus(ServicesList)
		}
		time.Sleep(2 * time.Second)
	}
}

func GetJSON() string {
	processLines := strings.Split(output, "\n")

	jsonArray := make([]string, 0)

	for _, line := range processLines {
		if line != "" {
			line = strings.Join(strings.Fields(line), " ")
			jsonArray = append(jsonArray,
				`"`+getServiceName(line)+`": {
					"is_active": "`+isActive(line)+`",
					"is_enabled": "`+isEnabled(line)+`"
				}
			`)
		}
	}

	obj := `{ "services_info": {` + strings.Join(jsonArray, ",") + `}}`

	ret := map[string]interface{}{}
	err := json.Unmarshal([]byte(obj), &ret)
	common.Error(err)

	return obj
}

func getProcessesStatus(services []string) (output string) {
	for _, service := range services {
		if service != "" {
			CommandServiceIsActive := config.GetStringSlice(Cfg, "on_runtime.commands.service_is_active")
			CommandServiceIsEnabled := config.GetStringSlice(Cfg, "on_runtime.commands.service_is_enabled")
			is_active := common.Cli(common.ReplaceStringInSlice(CommandServiceIsActive, "{service}", service))
			is_enabled := common.Cli(common.ReplaceStringInSlice(CommandServiceIsEnabled, "{service}", service))
			is_active = strings.ReplaceAll(is_active, "\n", "")
			is_enabled = strings.ReplaceAll(is_enabled, "\n", "")
			output += service + " " + is_active + " " + is_enabled + "\n"
		}
	}

	return output
}

func getServiceName(s string) string {
	return strings.Split(s, " ")[0]
}

func isActive(s string) string {
	arr := strings.Split(s, " ")
	if len(arr) < 3 {
		return "unknown"
	}
	return arr[1]
}

func isEnabled(s string) string {
	arr := strings.Split(s, " ")
	if len(arr) < 3 {
		return "unknown"
	}
	return arr[2]
}
