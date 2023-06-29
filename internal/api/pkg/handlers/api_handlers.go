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
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
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
	fmt.Fprintf(w, "%s", string(all.GetRawJSONs().GetJSON()))
}

// Playground for testing stuff.
func Playground(w http.ResponseWriter, r *http.Request) {
	L.Info("Playground", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(playground.Playground()))
}

// Model provides JSON from model name.
func Model(w http.ResponseWriter, r *http.Request) {
	L.Info("Model", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(model.GetJSON()))
}

// Cpu provides JSON from cpu.
func Cpu(w http.ResponseWriter, r *http.Request) {
	L.Info("Cpu", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(cpu.GetJSON()))
}

// Memory provides JSON from memory.
func Memory(w http.ResponseWriter, r *http.Request) {
	L.Info("Memory", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(memory.GetJSON()))
}

// Process provides JSON from processes.
func Process(w http.ResponseWriter, r *http.Request) {
	L.Info("Process", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(processes.GetJSON()))
}

// Storages provides JSON from storages.
func Storages(w http.ResponseWriter, r *http.Request) {
	L.Info("Storages", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(storage.GetJSON()))
}

// Services provides JSON from services.
func Services(w http.ResponseWriter, r *http.Request) {
	L.Info("Services", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(services.GetJSON()))
}

// Network provides JSON from network.
func Network(w http.ResponseWriter, r *http.Request) {
	L.Info("Network", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(network.GetJSON()))
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
