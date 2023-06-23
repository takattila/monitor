package main

import (
	"github.com/go-chi/chi"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/handlers"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/servers"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var s *settings.Settings

func init() {
	s = settings.New(common.GetConfigPath("api"))
	s.AutoReload()

	s.Data.Set("Memory", false)
	s.Data.Set("Services", false)
	s.Data.Set("TopProcesses", false)
	s.Data.Set("NetworkTraffic", false)
	s.Data.Set("Storage", false)

	handlers.Cfg, cpu.Cfg, memory.Cfg, model.Cfg, network.Cfg, processes.Cfg, services.Cfg, storage.Cfg = s, s, s, s, s, s, s, s

	go services.Watcher()
	go network.Stats()
}

func main() {
	router := chi.NewRouter()

	router.Get(config.GetString(s, "on_start.routes.all"), handlers.All)
	router.Get(config.GetString(s, "on_start.routes.playground"), handlers.Playground)
	router.Get(config.GetString(s, "on_start.routes.model"), handlers.Model)
	router.Get(config.GetString(s, "on_start.routes.cpu"), handlers.Cpu)
	router.Get(config.GetString(s, "on_start.routes.memory"), handlers.Memory)
	router.Get(config.GetString(s, "on_start.routes.processes"), handlers.Process)
	router.Get(config.GetString(s, "on_start.routes.storages"), handlers.Storages)
	router.Get(config.GetString(s, "on_start.routes.services"), handlers.Services)
	router.Get(config.GetString(s, "on_start.routes.network"), handlers.Network)
	router.Get(config.GetString(s, "on_start.routes.toggle"), handlers.Toggle)

	servers.ServeHTTP(config.GetInt(s, "on_start.port"), router)
}
