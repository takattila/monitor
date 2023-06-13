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
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

var Cfg *settings.Settings

// All ...
func All(w http.ResponseWriter, r *http.Request) {
	common.Info("All", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(all.GetJSON()))
}

// Playground ...
func Playground(w http.ResponseWriter, r *http.Request) {
	common.Info("Playground", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(playground.Playground()))
}

// Model ...
func Model(w http.ResponseWriter, r *http.Request) {
	common.Info("Model", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(model.GetJSON()))
}

// Cpu ...
func Cpu(w http.ResponseWriter, r *http.Request) {
	common.Info("Cpu", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(cpu.GetJSON()))
}

// Memory ...
func Memory(w http.ResponseWriter, r *http.Request) {
	common.Info("Memory", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(memory.GetJSON()))
}

// Process ...
func Process(w http.ResponseWriter, r *http.Request) {
	common.Info("Process", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(processes.GetJSON()))
}

// Storages ...
func Storages(w http.ResponseWriter, r *http.Request) {
	common.Info("Storages", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(storage.GetJSON()))
}

// Services ...
func Services(w http.ResponseWriter, r *http.Request) {
	common.Info("Services", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(services.GetJSON()))
}

// Network ...
func Network(w http.ResponseWriter, r *http.Request) {
	common.Info("Network", "Request IP:", r.RemoteAddr)
	fmt.Fprintf(w, "%s", string(network.GetJSON()))
}

// Toggle ...
func Toggle(w http.ResponseWriter, r *http.Request) {
	section := chi.URLParam(r, "section")
	status := chi.URLParam(r, "status")
	common.Info("section:", section, "status:", status)

	if section == "Memory" ||
		section == "Services" ||
		section == "TopProcesses" ||
		section == "NetworkTraffic" ||
		section == "Storage" {

		Cfg.Data.Set(section, status)

		fmt.Fprintf(w, `{"%s":"%s"}`, section, status)
	}
}
