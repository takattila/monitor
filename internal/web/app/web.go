package main

import (
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/internal/web/pkg/auth"
	"github.com/takattila/monitor/internal/web/pkg/handlers"
	"github.com/takattila/monitor/internal/web/pkg/servers"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	dir string
	s   *settings.Settings
	l   logger.Logger
)

func init() {
	dir = common.GetProgramDir()

	s = settings.New(common.GetConfigPath("web"))
	s.AutoReload()

	if err := auth.SaveCredentials((filepath.Join(dir, config.GetString(s, "on_start.auth_file"))), config.GetBool(s, "on_start.save_credentials")); err != nil {
		l.Fatal(err)
	}
}

func main() {
	router := chi.NewRouter()

	l := logger.New(
		config.GetLogLevel(s, "on_start.logger.level"),
		config.GetLogColor(s, "on_start.logger.color"),
	)

	h := handlers.Handler{
		ProgramDir:    dir,
		FilesDir:      config.GetString(s, "on_start.web_sources_directory"),
		AuthFile:      config.GetString(s, "on_start.auth_file"),
		LoginPage:     config.GetString(s, "on_start.pages.login"),
		InternalPage:  config.GetString(s, "on_start.pages.internal"),
		LoginRoute:    config.GetString(s, "on_start.routes.login"),
		InternalRoute: config.GetString(s, "on_start.routes.internal"),
		Cfg:           s,
		L:             l,
	}

	router.HandleFunc(config.GetString(s, "on_start.routes.index"), h.Index)
	router.Get(config.GetString(s, "on_start.routes.login"), h.Login)
	router.Get(config.GetString(s, "on_start.routes.logout"), h.Logout)
	router.Get(config.GetString(s, "on_start.routes.internal"), h.Internal)
	router.Get(config.GetString(s, "on_start.routes.api"), h.Api)
	router.Get(config.GetString(s, "on_start.routes.toggle"), h.Toggle)
	router.Post(config.GetString(s, "on_start.routes.systemctl"), h.SystemCtl)
	router.Post(config.GetString(s, "on_start.routes.power"), h.Power)

	s := servers.Server{
		Port:       config.GetInt(s, "on_start.port"),
		Domain:     config.GetString(s, "on_start.domain"),
		Router:     router,
		RoutePath:  config.GetString(s, "on_start.routes.web"),
		ProgramDir: dir,
		FilesDir:   config.GetString(s, "on_start.web_sources_directory"),
		Cfg:        s,
		L:          l,
	}

	s.Files()
	s.Start()
}
