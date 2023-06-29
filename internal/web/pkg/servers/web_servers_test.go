package servers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/internal/common/pkg/config"
	"github.com/takattila/monitor/internal/web/pkg/handlers"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	WebServersSuite struct {
		suite.Suite
	}
)

var (
	gitRootPath = strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	s           = getConfig("web", "linux")
	h           = handlers.Handler{
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

func (a WebServersSuite) TestStartHTTP() {
	apiport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	webport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	go startWebServer(a.T(), apiport, webport)
	time.Sleep(100 * time.Millisecond)

	internalURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.internal"))
	resp, err := req("GET", internalURL, nil)
	a.Equal(nil, err)
	a.Equal(200, resp.StatusCode)
}

func (a WebServersSuite) TestStartTLS() {
	oldTlsPort := tlsPort
	defer func() { tlsPort = oldTlsPort }()

	oldHttpPort := httpPort
	defer func() { httpPort = oldHttpPort }()

	apiport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	webport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	httpPort, err = freeport.GetFreePort()
	a.Equal(nil, err)

	tlsPort = webport

	go startWebServer(a.T(), apiport, webport)
	time.Sleep(100 * time.Millisecond)

	internalURL := fmt.Sprintf("https://127.0.0.1:%d%s", webport, config.GetString(s, "on_start.routes.internal"))
	_, err = req("GET", internalURL, nil)
	a.Contains(fmt.Sprint(err), "remote error: tls: internal error")
}

func (a WebServersSuite) TestFilesNotAllowed() {
	oldFilesRute := s.Data.Get("on_start.routes.web")
	s.Data.Set("on_start.routes.web", "/monitor/web*")
	defer func() { s.Data.Set("on_start.routes.web", oldFilesRute) }()

	apiport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	webport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	go startWebServer(a.T(), apiport, webport)
	time.Sleep(100 * time.Millisecond)

	internalURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), config.GetString(s, "on_start.routes.web"))
	resp, err := req("GET", internalURL, nil)
	a.Equal(nil, err)
	a.Equal(404, resp.StatusCode)
}

func (a WebServersSuite) TestFilesRedirect301() {
	apiport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	webport, err := freeport.GetFreePort()
	a.Equal(nil, err)

	go startWebServer(a.T(), apiport, webport)
	time.Sleep(100 * time.Millisecond)

	internalURL := fmt.Sprintf("http://127.0.0.1:%d%s", config.GetInt(s, "on_start.port"), "/monitor/web")
	resp, err := req("GET", internalURL, nil)
	a.Equal(nil, err)
	a.Equal(404, resp.StatusCode)
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

func startWebServer(t *testing.T, apiport, webport int) {
	h.Cfg.Data.Set("on_runtime.api.port", apiport)
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

	s := Server{
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

func getConfig(service, system string) *settings.Settings {
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func TestWebServersSuite(t *testing.T) {
	suite.Run(t, new(WebServersSuite))
}
