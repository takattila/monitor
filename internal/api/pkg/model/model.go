package model

import (
	"encoding/json"
	"runtime"
	"strings"

	"github.com/matishsiao/goInfo"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
)

type Model struct {
	ModelName string `json:"model_name"`
}

func GetJSON() string {
	m := Model{}
	m.getModelName()
	b, err := json.Marshal(m)
	common.Error(err)

	return string(b)
}

func (m *Model) getModelName() *Model {
	ret, _ := Cfg.GetStringSlice("on_runtime.commands.model_name")
	if len(ret) == 0 {
		return m.getModelNameOS()
	}
	m.ModelName = strings.Replace(common.Cli(ret), "\n", "", -1)
	return m
}

func (m *Model) getModelNameOS() *Model {
	info, _ := goInfo.GetInfo()
	m.ModelName = strings.Title(info.Kernel + " " + info.Core + " " + runtime.GOARCH)
	return m
}
