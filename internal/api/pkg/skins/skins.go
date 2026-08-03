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

	excluded := []string{"progress-presets", "light-mode"}

	var skins []string

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".css" {
			skin := strings.ReplaceAll(file.Name(), ext, "")
			if !contains(excluded, skin) {
				skins = append(skins, `"`+skin+`"`)
			}
		}
	}

	L.Debug("skins", skins)
	obj := `{ "skins": [` + strings.Join(skins, ",") + `]}`

	ret := map[string]interface{}{}
	err = json.Unmarshal([]byte(obj), &ret)
	L.Error(err)

	return obj
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
