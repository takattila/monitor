package logos

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
	LogosPath = "./web/img"
)

// GetJSON returns with a JSON that holds information from available logos.
func GetJSON() string {
	files, err := ioutil.ReadDir(LogosPath)
	L.Error(err)

	var logos []string

	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".png" {
			logo := strings.ReplaceAll(file.Name(), ext, "")
			logos = append(logos, `"`+logo+`"`)
		}
	}

	L.Debug("logos", logos)
	obj := `{ "logos": [` + strings.Join(logos, ",") + `]}`

	ret := map[string]interface{}{}
	err = json.Unmarshal([]byte(obj), &ret)
	L.Error(err)

	return obj
}
