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
	Sleep  = 2 * time.Second

	// getProcessesStatus collects status information about a service: 'is_active' and 'is_enabled'.
	// Output example: 'service_name active enabled'.
	getProcessesStatus = func(services []string) (output string) {
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
)

// Watcher collects services data into output variable.
// It should be run in the background by starting with: 'go Watcher()'.
func Watcher() {
	ServicesList := config.GetStringSlice(Cfg, "on_runtime.services_list")
	output = getProcessesStatus(ServicesList)
	for {
		if Cfg.Data.GetBool("Services") {
			ServicesList := config.GetStringSlice(Cfg, "on_runtime.services_list")
			output = getProcessesStatus(ServicesList)
		}
		time.Sleep(Sleep)
	}
}

// GetJSON returns with a JSON that holds information from services: 'is_active' and 'is_enabled'.
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

// getServiceName fetches the service name from a string.
func getServiceName(s string) string {
	return strings.Split(s, " ")[0]
}

// isActive fetches the 'active' state from a string.
func isActive(s string) string {
	arr := strings.Split(s, " ")
	if len(arr) < 3 {
		return "unknown"
	}
	return arr[1]
}

// isEnabled fetches 'enabled' state from a string.
func isEnabled(s string) string {
	arr := strings.Split(s, " ")
	if len(arr) < 3 {
		return "unknown"
	}
	return arr[2]
}
