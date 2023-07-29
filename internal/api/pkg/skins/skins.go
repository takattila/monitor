package skins

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg       *settings.Settings
	L         logger.Logger
	SkinsPath = "./web/css"
)

// GetJSON returns with a JSON that holds information from available skins.
func GetJSON() string {
	files, err := ioutil.ReadDir(SkinsPath)
	L.Error(err)

	var skins []string

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".css" {
			skin := strings.ReplaceAll(file.Name(), ext, "")
			skins = append(skins, `"`+skin+`"`)
		}
	}

	L.Debug("skins", skins)
	obj := `{ "skins": [` + strings.Join(skins, ",") + `]}`

	ret := map[string]interface{}{}
	err = json.Unmarshal([]byte(obj), &ret)
	L.Error(err)

	return obj
}
