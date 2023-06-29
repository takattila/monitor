package main

import (
	"fmt"

	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/internal/web/pkg/auth"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

func main() {
	l := logger.New(logger.ErrorLevel, logger.ColorOn)

	dir := common.GetProgramDir()
	s := settings.New(common.GetConfigPath("web"))
	authFile := dir + config.GetString(s, "on_start.auth_file")
	err := auth.SaveCredentials(authFile, true)
	l.Fatal(err)

	println()
	fmt.Println("Credentials saved into:", authFile)
}
