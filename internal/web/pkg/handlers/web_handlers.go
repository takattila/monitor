package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/internal/web/pkg/auth"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

type (
	Handler struct {
		ProgramDir    string
		FilesDir      string
		AuthFile      string
		AllowedIP     string
		ApiService    ApiService
		LoginPage     string
		InternalPage  string
		LoginRoute    string
		InternalRoute string
		Cfg           *settings.Settings
	}

	ApiService struct {
		Url  string
		Port int
	}
)

// Internal serves statistics page.
func (h *Handler) Internal(w http.ResponseWriter, r *http.Request) {
	userName := auth.GetUserName(r)
	common.Info("userName:", userName)
	if userName == "" {
		http.Redirect(w, r, h.LoginRoute, 302)
	}

	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		t := time.Now()

		tmpl := template.Must(
			template.ParseFiles(
				filepath.Join(
					h.ProgramDir,
					h.FilesDir,
					h.InternalPage)))

		data := struct {
			Version         string
			Skin            string
			Logo            string
			RouteSystemCtl  string
			RoutePower      string
			RouteToggle     string
			RouteLogout     string
			RouteApi        string
			RouteIndex      string
			RouteWebPath    string
			IntervalSeconds int
		}{
			Version:         fmt.Sprint(t.Year()) + fmt.Sprint(int(t.Month())) + fmt.Sprint(t.YearDay()) + fmt.Sprint(t.Minute()) + fmt.Sprint(t.Second()) + fmt.Sprint(t.Nanosecond()),
			Skin:            config.GetString(h.Cfg, "on_runtime.theme.skin"),
			Logo:            config.GetString(h.Cfg, "on_runtime.theme.logo"),
			RouteSystemCtl:  config.GetString(h.Cfg, "on_start.routes.systemctl"),
			RoutePower:      config.GetString(h.Cfg, "on_start.routes.power"),
			RouteToggle:     config.GetString(h.Cfg, "on_start.routes.toggle"),
			RouteLogout:     config.GetString(h.Cfg, "on_start.routes.logout"),
			RouteApi:        config.GetString(h.Cfg, "on_start.routes.api"),
			RouteIndex:      config.GetString(h.Cfg, "on_start.routes.index"),
			RouteWebPath:    config.GetString(h.Cfg, "on_start.routes.web"),
			IntervalSeconds: config.GetInt(h.Cfg, "on_runtime.interval_seconds"),
		}

		tmpl.Execute(w, data)
	}
}

// Login serves a login page.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		t := time.Now()

		tmpl := template.Must(
			template.ParseFiles(
				filepath.Join(
					h.ProgramDir,
					h.FilesDir,
					h.LoginPage)))

		data := struct {
			Version      string
			Skin         string
			Logo         string
			RouteIndex   string
			RouteWebPath string
		}{
			Version:      fmt.Sprint(t.Year()) + fmt.Sprint(int(t.Month())) + fmt.Sprint(t.YearDay()) + fmt.Sprint(t.Minute()) + fmt.Sprint(t.Second()) + fmt.Sprint(t.Nanosecond()),
			Skin:         config.GetString(h.Cfg, "on_runtime.theme.skin"),
			Logo:         config.GetString(h.Cfg, "on_runtime.theme.logo"),
			RouteIndex:   config.GetString(h.Cfg, "on_start.routes.index"),
			RouteWebPath: config.GetString(h.Cfg, "on_start.routes.web"),
		}

		tmpl.Execute(w, data)
	}
}

// SystemCtl queries or sends control commands to the systemd manager.
func (h *Handler) SystemCtl(w http.ResponseWriter, r *http.Request) {
	userName := auth.GetUserName(r)
	common.Info("userName:", userName)
	if userName == "" {
		http.Redirect(w, r, h.LoginRoute, 302)
	}

	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		action := chi.URLParam(r, "action")
		service := chi.URLParam(r, "service")

		if common.SliceContains([]string{"start", "stop", "restart", "enable", "disable"}, action) {
			common.Info("action:", action, "service:", service)
			cmd := config.GetStringSlice(h.Cfg, "on_runtime.commands.systemctl")
			cmd = common.ReplaceStringInSlice(cmd, "{action}", action)
			cmd = common.ReplaceStringInSlice(cmd, "{service}", service)
			fmt.Fprintf(w, "%s", common.Cli(cmd))
		} else {
			common.Error(fmt.Errorf("action: %s is not allowed", action))
		}
	}
}

// Power runs power actions: shutdown or reboot.
func (h *Handler) Power(w http.ResponseWriter, r *http.Request) {
	userName := auth.GetUserName(r)
	common.Info("userName:", userName)
	if userName == "" {
		http.Redirect(w, r, h.LoginRoute, 302)
	}

	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		action := chi.URLParam(r, "action")

		initNumber := "0"
		if action == "reboot" {
			initNumber = "6"
		}

		common.Info("action:", action)
		cmd := config.GetStringSlice(h.Cfg, "on_runtime.commands.init")
		cmd = common.ReplaceStringInSlice(cmd, "{number}", initNumber)
		fmt.Fprintf(w, "%s", common.Cli(cmd))
	}
}

// Api handler sends requests to the MONITOR-API service.
// Basicaly it is a proxy.
func (h *Handler) Api(w http.ResponseWriter, r *http.Request) {
	userName := auth.GetUserName(r)
	common.Info("userName:", userName)
	if userName == "" {
		http.Redirect(w, r, h.LoginRoute, 302)
	}

	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		statistics := chi.URLParam(r, "statistics")

		requestURL := fmt.Sprintf("%s:%d/%s", config.GetString(h.Cfg, "on_runtime.api.url"), config.GetInt(h.Cfg, "on_runtime.api.port"), statistics)
		res, err := http.Get(requestURL)
		if err != nil {
			common.Error(fmt.Errorf("making http request: %v", err))
			return
		}

		common.Info(requestURL, "client: status code:", res.StatusCode)

		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			common.Error(fmt.Errorf("client: could not read response body: %v", err))
		}
		fmt.Fprintf(w, "%s", resBody)
	}
}

// Index checks user credentials.
func (h *Handler) Index(response http.ResponseWriter, request *http.Request) {
	common.Info("AllowedIP:", config.GetString(h.Cfg, "on_runtime.allowed_ip"), "Request IP:", request.RemoteAddr)
	name := request.FormValue("uname")
	pass := request.FormValue("psw")
	redirectTarget := h.LoginRoute
	authenticated := auth.Authenticate(h.ProgramDir+h.AuthFile, name, pass)
	common.Info("Authenticate", authenticated)

	if name != "" && pass != "" && authenticated {
		auth.SetSession(name, response)
		redirectTarget = h.InternalRoute
	}
	http.Redirect(response, request, redirectTarget, 302)
}

// Logout clears user session.
func (h *Handler) Logout(response http.ResponseWriter, request *http.Request) {
	auth.ClearSession(response)
	http.Redirect(response, request, h.LoginRoute, 302)
}

// Toggle ...
func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	userName := auth.GetUserName(r)
	common.Info("userName:", userName)
	if userName == "" {
		http.Redirect(w, r, h.LoginRoute, 302)
	}

	if IPisAllowed(r.RemoteAddr, config.GetString(h.Cfg, "on_runtime.allowed_ip")) {
		section := chi.URLParam(r, "section")
		status := chi.URLParam(r, "status")

		common.Info("section:", section, "status:", status)

		requestURL := fmt.Sprintf("%s:%d/toggle/%s/%s",
			config.GetString(h.Cfg, "on_runtime.api.url"),
			config.GetInt(h.Cfg, "on_runtime.api.port"),
			section,
			status)

		res, err := http.Get(requestURL)
		if err != nil {
			common.Error(fmt.Errorf("making http request: %v", err))
			return
		}

		common.Info(requestURL, "client: status code:", res.StatusCode)

		resBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			common.Error(fmt.Errorf("client: could not read response body: %v", err))
		}
		fmt.Fprintf(w, "%s", resBody)
	}
}

// IPisAllowed checks whether the request IP is allowed or not.
func IPisAllowed(requestIP, allowedIP string) bool {
	common.Info("Request IP:", requestIP, "Allowed IP:", allowedIP)
	// If the IP contains port number as well, the port should be rmoved from the requestIP.
	if strings.Contains(requestIP, ":") {
		requestIP = strings.Split(requestIP, ":")[0]
	}
	if allowedIP == "0.0.0.0" {
		common.Info("Allowed IP was not set")
		return true
	}
	if strings.Contains(allowedIP, ",") {
		common.Info("Multiple IP were set for allowedIP:", allowedIP)
		ips := strings.Split(allowedIP, ",")
		ret := false
		for _, ip := range ips {
			common.Info("allowedIP", ip, "requestIP", requestIP)
			if requestIP == ip {
				ret = true
			}
		}
		common.Info("IP is allowed ("+requestIP+"):", ret)
		return ret
	}
	if allowedIP != requestIP {
		common.Error(fmt.Errorf("IP not allowed: " + requestIP))
		return false
	}
	return true
}
