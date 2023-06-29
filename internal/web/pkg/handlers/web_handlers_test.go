package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/handlers"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/playground"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	apiservers "github.com/takattila/monitor/internal/api/pkg/servers"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/internal/web/pkg/servers"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	WebHandlersSuite struct {
		suite.Suite
	}
)

var (
	gitRootPath = strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	s           = getConfig("web", "linux")
	h           = &Handler{
		ProgramDir:    gitRootPath,
		FilesDir:      config.GetString(s, "on_start.web_sources_directory"),
		AuthFile:      config.GetString(s, "on_start.auth_file"),
		LoginPage:     config.GetString(s, "on_start.pages.login"),
		InternalPage:  config.GetString(s, "on_start.pages.internal"),
		LoginRoute:    config.GetString(s, "on_start.routes.login"),
		InternalRoute: config.GetString(s, "on_start.routes.internal"),
		Cfg:           s,
		L:             logger.New(logger.NoneLevel, logger.ColorOff),
	}

	r = chi.NewRouter()
)

func (a WebHandlersSuite) TestInternalNotAuthenticated() {
	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	internalURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.internal"))
	resp, err := req("GET", internalURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestLoginOk() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	loginURL = fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err = req("GET", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestSystemCtlOk() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	oldSystemCtlCmd := h.Cfg.Data.Get("on_runtime.commands.systemctl")
	h.Cfg.Data.Set("on_runtime.commands.systemctl", []string{"bash", "-c", "echo ok"})
	defer func() { h.Cfg.Data.Set("on_runtime.commands.systemctl", oldSystemCtlCmd) }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	systemctlURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/systemctl/start/myservice")
	resp, err = req("POST", systemctlURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestSystemCtlNotAuthenticated() {
	systemctlURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/systemctl/start/myservice")
	resp, err := req("POST", systemctlURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestSystemCtlActionNotAllowed() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	systemctlURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/systemctl/bad_action/myservice")
	resp, err = req("POST", systemctlURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestPowerOk() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	powerURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/power/reboot")
	resp, err = req("POST", powerURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestPowerNotAuthenticated() {
	powerURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/power/reboot")
	resp, err := req("POST", powerURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestApiOk() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	go startApiServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	apiURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/api/cpu")
	resp, err = req("GET", apiURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestApiApiNotFound() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	apiURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/api/cpu")
	resp, err = req("GET", apiURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestApiNotAuthenticated() {
	apiURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/api/cpu")
	resp, err := req("GET", apiURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestLogoutOk() {
	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	logoutURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.logout"))
	resp, err := req("GET", logoutURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestToggleOk() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	go startApiServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	toggleURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/toggle/Memory/true")
	resp, err = req("GET", toggleURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestToggleNotAuthenticated() {
	toggleURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/toggle/Memory/true")
	resp, err := req("GET", toggleURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestToggleApiNotFound() {
	user := "username"
	pass := "password"

	authdb := newTestAuthDB(a.T(), user, pass)
	defer func() { _ = os.Remove(authdb) }()

	oldGetUsernameFunc := bypassGetUsername(user)
	defer func() { getUsername = oldGetUsernameFunc }()

	go startWebServer(a.T())
	time.Sleep(100 * time.Millisecond)

	form := url.Values{}
	form.Add("uname", user)
	form.Add("psw", pass)

	loginURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.index"))
	resp, err := req("POST", loginURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)

	toggleURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/toggle/Memory/true")
	resp, err = req("GET", toggleURL, strings.NewReader(form.Encode()))
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebHandlersSuite) TestIPisAllowedIPNotSet() {
	allowed := IPisAllowed("127.0.0.1", "0.0.0.0", h)
	a.Equal(true, allowed)
}

func (a WebHandlersSuite) TestIPisAllowedOK() {
	allowed := IPisAllowed("127.0.0.1", "127.0.0.1", h)
	a.Equal(true, allowed)
}

func (a WebHandlersSuite) TestIPisAllowedMultipleIPs() {
	allowed := IPisAllowed("127.0.0.1", "10.1.1.100,127.0.0.1,192.18.0.1", h)
	a.Equal(true, allowed)
}

func (a WebHandlersSuite) TestIPisAllowedNotAllowedIP() {
	allowed := IPisAllowed("127.0.0.1", "192.18.0.1", h)
	a.Equal(false, allowed)
}

func startWebServer(t *testing.T) {
	apiport, err := freeport.GetFreePort()
	if err != nil {
		t.Errorf("[ERROR] freeport.GetFreePort: %s\n", err)
	}
	h.Cfg.Data.Set("on_runtime.api.port", apiport)

	webport, err := freeport.GetFreePort()
	if err != nil {
		t.Errorf("[ERROR] freeport.GetFreePort: %s\n", err)
	}
	h.Cfg.Data.Set("on_start.port", webport)
	h.Cfg.Data.Set("on_runtime.commands.init", []string{"bash", "-c", "echo reboot"})

	r.HandleFunc(config.GetString(s, "on_start.routes.index"), h.Index)
	r.Get(config.GetString(s, "on_start.routes.login"), h.Login)
	r.Get(config.GetString(s, "on_start.routes.logout"), h.Logout)
	r.Get(config.GetString(s, "on_start.routes.internal"), h.Internal)
	r.Get(config.GetString(s, "on_start.routes.api"), h.Api)
	r.Get(config.GetString(s, "on_start.routes.toggle"), h.Toggle)
	r.Post(config.GetString(s, "on_start.routes.systemctl"), h.SystemCtl)
	r.Post(config.GetString(s, "on_start.routes.power"), h.Power)

	s := servers.Server{
		Port:       config.GetInt(s, "on_start.port"),
		Domain:     config.GetString(s, "on_start.domain"),
		Router:     r,
		RoutePath:  config.GetString(s, "on_start.routes.web"),
		ProgramDir: h.FilesDir,
		FilesDir:   config.GetString(s, "on_start.web_sources_directory"),
		Cfg:        s,
		L:          logger.New(logger.NoneLevel, logger.ColorOff),
	}

	s.Files()
	s.Start()
}

func startApiServer(t *testing.T) {
	s := getConfig("api", "linux")
	s.Data.Set("on_start.port", h.Cfg.Data.Get("on_runtime.api.port"))

	s.Data.Set("Memory", false)
	s.Data.Set("Services", false)
	s.Data.Set("TopProcesses", false)
	s.Data.Set("NetworkTraffic", false)
	s.Data.Set("Storage", false)

	handlers.Cfg, cpu.Cfg, memory.Cfg, model.Cfg, network.Cfg, processes.Cfg, services.Cfg, storage.Cfg = s, s, s, s, s, s, s, s

	l := logger.New(logger.NoneLevel, logger.ColorOff)
	cpu.L, handlers.L, memory.L, model.L, network.L, playground.L, processes.L, apiservers.L, services.L, storage.L = l, l, l, l, l, l, l, l, l, l

	go services.Watcher()
	go network.Stats()

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

	apiservers.ServeHTTP(config.GetInt(s, "on_start.port"), router)
}

func req(method, url string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %v", err)
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %v", err)
	}
	defer res.Body.Close()

	return res, nil
}

func bypassGetUsername(username string) func(r *http.Request) string {
	oldGetUsernameFunc := getUsername
	getUsername = func(r *http.Request) string {
		return username
	}
	return oldGetUsernameFunc
}

func newTestAuthDB(t *testing.T, user, pass string) string {
	h.AuthFile = "/configs/testauth.db"
	authdbFullPath := h.ProgramDir + h.AuthFile

	f, err := os.OpenFile(authdbFullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		t.Fatalf("os.OpenFile: %v", err)
	}
	defer f.Close()

	authString := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	if _, err := f.WriteString(authString + "\n"); err != nil {
		t.Fatalf("f.WriteString: %v", err)
	}

	return authdbFullPath
}

func getConfig(service, system string) *settings.Settings {
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestWebHandlersSuite(t *testing.T) {
	suite.Run(t, new(WebHandlersSuite))
}
