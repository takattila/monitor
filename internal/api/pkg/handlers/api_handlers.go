package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/internal/api/pkg/all"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/playground"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/run"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/skins"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

var (
	Cfg *settings.Settings
	L   logger.Logger
)

// All provides JSON from all sections.
func All(w http.ResponseWriter, r *http.Request) {
	L.Debug("All", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", all.GetRawJSONs().GetJSON())
}

// Playground for testing stuff.
func Playground(w http.ResponseWriter, r *http.Request) {
	L.Info("Playground", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", playground.Playground())
}

// Model provides JSON from model name.
func Model(w http.ResponseWriter, r *http.Request) {
	L.Info("Model", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", model.GetJSON())
}

// Cpu provides JSON from cpu.
func Cpu(w http.ResponseWriter, r *http.Request) {
	L.Info("Cpu", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", cpu.GetJSON())
}

// Memory provides JSON from memory.
func Memory(w http.ResponseWriter, r *http.Request) {
	L.Info("Memory", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", memory.GetJSON())
}

// Process provides JSON from processes.
func Process(w http.ResponseWriter, r *http.Request) {
	L.Info("Process", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", processes.GetJSON())
}

// Storages provides JSON from storages.
func Storages(w http.ResponseWriter, r *http.Request) {
	L.Info("Storages", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", storage.GetJSON())
}

// Services provides JSON from services.
func Services(w http.ResponseWriter, r *http.Request) {
	L.Info("Services", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", services.GetJSON())
}

// Network provides JSON from network.
func Network(w http.ResponseWriter, r *http.Request) {
	L.Info("Network", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", network.GetJSON())
}

// Toggle turns a specific JSON provider on or off.
func Toggle(w http.ResponseWriter, r *http.Request) {
	section := chi.URLParam(r, "section")
	status := chi.URLParam(r, "status")
	L.Info("section:", section, "status:", status)

	if section == "Memory" ||
		section == "Services" ||
		section == "TopProcesses" ||
		section == "NetworkTraffic" ||
		section == "Storage" {

		Cfg.Data.Set(section, status)

		fmt.Fprintf(w, `{"%s":"%s"}`, section, status)
	}
}

// RunList provides a JSON from the commands can be run.
func RunList(w http.ResponseWriter, r *http.Request) {
	L.Info("RunList", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", run.GetJSON())
}

// RunExec executes a specific command by its name.
func RunExec(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	L.Info("RunExec", "Request IP:", r.RemoteAddr)
	if name != "" {
		go func() {
			err := run.Exec(name)
			L.Error(err)
			L.Warning(common.GetProgramDir())
		}()
	}
	fmt.Fprintf(w, `{ "started": "%s" }`, name)
}

// RunStdOut returns with the specific run's output.
func RunStdOut(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	L.Info("RunStdOut", "Request IP:", r.RemoteAddr)
	output := run.StdOut(name)
	fmt.Fprintf(w, `%s`, output)
}

// Skins returns with a list of skins.
func Skins(w http.ResponseWriter, r *http.Request) {
	L.Info("Skins", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, `%s`, skins.GetJSON())
}
