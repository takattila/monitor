package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/internal/api/pkg/cpu"
	"github.com/takattila/monitor/internal/api/pkg/memory"
	"github.com/takattila/monitor/internal/api/pkg/model"
	"github.com/takattila/monitor/internal/api/pkg/network"
	"github.com/takattila/monitor/internal/api/pkg/playground"
	"github.com/takattila/monitor/internal/api/pkg/processes"
	"github.com/takattila/monitor/internal/api/pkg/run"
	"github.com/takattila/monitor/internal/api/pkg/servers"
	"github.com/takattila/monitor/internal/api/pkg/services"
	"github.com/takattila/monitor/internal/api/pkg/storage"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

type (
	ApiHandlersSuite struct {
		suite.Suite
	}
	response struct {
		httpresponse *http.Response
		responsebody string
		status       int
		error        error
	}
)

func (a ApiHandlersSuite) TestAll() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", false)
	s.Data.Set("Services", false)
	s.Data.Set("TopProcesses", false)
	s.Data.Set("NetworkTraffic", false)
	s.Data.Set("Storage", false)

	cpu.Cfg, memory.Cfg, model.Cfg, network.Cfg, processes.Cfg, run.Cfg, services.Cfg, storage.Cfg = s, s, s, s, s, s, s, s

	l := logger.New(logger.NoneLevel, logger.ColorOff)
	cpu.L, L, memory.L, model.L, network.L, playground.L, processes.L, run.L, servers.L, services.L, storage.L = l, l, l, l, l, l, l, l, l, l, l

	r := chi.NewRouter()
	r.Get("/all", All)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/all", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "model_name")
	a.Contains(request.responsebody, "processor_info")
	a.Contains(request.responsebody, "storage_info")
	a.Contains(request.responsebody, "process_info")
	a.Contains(request.responsebody, "services_info")
	a.Contains(request.responsebody, "network_info")
}

func (a ApiHandlersSuite) TestPlayground() {
	r := chi.NewRouter()
	r.Get("/playground", Playground)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/playground", nil)

	a.Equal(200, request.status)
	a.NotNil(request.responsebody)
}

func (a ApiHandlersSuite) TestModel() {
	s := getConfig("api", "linux")
	model.Cfg = s

	r := chi.NewRouter()
	r.Get("/model", Model)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/model", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "model_name")
}

func (a ApiHandlersSuite) TestCpu() {
	s := getConfig("api", "linux")
	cpu.Cfg = s

	r := chi.NewRouter()
	r.Get("/cpu", Cpu)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/cpu", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "processor_info")
}

func (a ApiHandlersSuite) TestMemory() {
	s := getConfig("api", "linux")
	memory.Cfg = s

	r := chi.NewRouter()
	r.Get("/memory", Memory)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/memory", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "memory_info")
}

func (a ApiHandlersSuite) TestProcess() {
	s := getConfig("api", "linux")
	processes.Cfg = s

	r := chi.NewRouter()
	r.Get("/process", Process)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/process", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "process_info")
}

func (a ApiHandlersSuite) TestStorages() {
	s := getConfig("api", "linux")
	storage.Cfg = s

	r := chi.NewRouter()
	r.Get("/storage", Storages)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/storage", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "storage_info")
}

func (a ApiHandlersSuite) TestServices() {
	s := getConfig("api", "linux")
	services.Cfg = s

	r := chi.NewRouter()
	r.Get("/services", Services)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/services", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "services_info")
}

func (a ApiHandlersSuite) TestNetwork() {
	s := getConfig("api", "linux")
	network.Cfg = s

	r := chi.NewRouter()
	r.Get("/network", Network)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/network", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "network_info")
}

func (a ApiHandlersSuite) TestToggle() {
	s := getConfig("api", "linux")
	s.Data.Set("Memory", false)
	Cfg = s

	r := chi.NewRouter()
	r.Get("/toggle/{section}/{status}", Toggle)

	ts := httptest.NewServer(r)
	defer ts.Close()
	request := request(ts, "GET", "/toggle/Memory/true", nil)

	a.Equal(200, request.status)
	a.Contains(request.responsebody, "Memory")
	a.Contains(request.responsebody, "true")
}

func getConfig(service, system string) *settings.Settings {
	gitRootPath := strings.ReplaceAll(common.Cli([]string{"bash", "-c", "git rev-parse --show-toplevel"}), "\n", "")
	configPath := gitRootPath + "/configs/" + service + "." + system + ".yaml"
	s := settings.New(configPath)
	s.AutoReload()
	return s
}

func request(ts *httptest.Server, method, path string, body io.Reader) response {
	request, requesterror := http.NewRequest(method, ts.URL+path, body)
	clientresponse, defaultclienterror := http.DefaultClient.Do(request)
	respBody, rerr := ioutil.ReadAll(clientresponse.Body)
	defer func() { _ = clientresponse.Body.Close() }()
	return response{
		httpresponse: clientresponse,
		responsebody: string(respBody),
		status:       clientresponse.StatusCode,
		error:        fmt.Errorf("http.NewRequest: %s, DefaultClient.Do: %s, ioutil.ReadAll: %s", requesterror, defaultclienterror, rerr),
	}
}

func TestApiHandlersSuite(t *testing.T) {
	suite.Run(t, new(ApiHandlersSuite))
}
