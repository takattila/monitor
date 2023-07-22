package run

import (
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg       *settings.Settings
	L         logger.Logger
	CmdFolder = "./cmd/"
)
