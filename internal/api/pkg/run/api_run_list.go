package run

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetJSON provides a JSON representation of the list of existing commands that can be run.
func GetJSON() string {
	var commands []string

	for command, _ := range Cfg.Data.GetStringMapString("on_runtime.run") {
		slice := Cfg.Data.GetStringSlice("on_runtime.run." + command)
		item := strings.Join(slice, ` `)
		b, _ := json.Marshal(item)
		commands = append(commands, `"`+command+`":`+string(b))
	}

	var ret string

	if len(commands) > 0 {
		ret = `{ "run_list": {` + strings.Join(commands, ",") + `} }`
	} else {
		ret = `{ "run_list": {} }`
		L.Error(fmt.Errorf("the lenght of the 'on_runtime.run' is: %d", len(commands)))
	}

	return ret
}
